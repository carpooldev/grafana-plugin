package plugin

import (
	"container/heap"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/valyala/fasthttp"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

// Make sure Datasource implements required interfaces. This is important to do
// since otherwise we will only get a not implemented error response from plugin in
// runtime. In this example datasource instance implements backend.QueryDataHandler,
// backend.CheckHealthHandler interfaces. Plugin should not implement all these
// interfaces- only those which are required for a particular task.
var (
	_ backend.QueryDataHandler      = (*Datasource)(nil)
	_ backend.CheckHealthHandler    = (*Datasource)(nil)
	_ instancemgmt.InstanceDisposer = (*Datasource)(nil)
)

type MetricType string

const (
	Invocations              MetricType = "invocations"
	UniqueSigners            MetricType = "uniqueSigners"
	Failures                 MetricType = "failures"
	FailureRate              MetricType = "failureRate"
	ProgramDeployments       MetricType = "programDeployments"
	FailedProgramDeployments MetricType = "failedProgramDeployments"
	TopInstructions          MetricType = "topInstructions"
)

func calculateBuckets(start, end time.Time, bucketSeconds int32) int32 {
	buckets := int32(end.Sub(start).Seconds()) / bucketSeconds
	if buckets < 1 {
		buckets = 1
	}
	return buckets
}

func clampBucketSeconds(start, end time.Time, bucketSeconds int32, maxBuckets int32) int32 {
	buckets := calculateBuckets(start, end, bucketSeconds)
	if buckets > maxBuckets {
		return (int32(end.Sub(start).Seconds()) / 256) + 1
	}
	return bucketSeconds
}

func getQueryType(qt MetricType) MetricType {
	switch qt {
	case Invocations:
		return Invocations
	case UniqueSigners:
		return UniqueSigners
	case Failures:
		return Invocations
	case FailureRate:
		return Invocations
	case ProgramDeployments:
		return ProgramDeployments
	case FailedProgramDeployments:
		return FailedProgramDeployments
	case TopInstructions:
		return Invocations
	default:
		return Invocations
	}
}

// NewDatasource creates a new datasource instance.
func NewDatasource(inst backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	readTimeout, _ := time.ParseDuration("15000ms")
	writeTimeout, _ := time.ParseDuration("15000ms")
	maxIdleConnDuration, _ := time.ParseDuration("1h")
	var client = &fasthttp.Client{
		ReadTimeout:              readTimeout,
		WriteTimeout:             writeTimeout,
		MaxIdleConnDuration:      maxIdleConnDuration,
		NoDefaultUserAgentHeader: true,
		DisablePathNormalizing:   true,
		Dial: (&fasthttp.TCPDialer{
			Concurrency:      4096,
			DNSCacheDuration: time.Hour,
		}).Dial,
	}
	log.DefaultLogger.Info("NewDatasource", "inst", inst)
	var settings datasourceSettings
	if err := json.Unmarshal(inst.JSONData, &settings); err != nil {
		return nil, err
	}
	return &Datasource{
		Client:     client,
		host:       settings.Url,
		maxBuckets: settings.MaxBuckets,
		apiKey:     inst.DecryptedSecureJSONData["apiKey"],
	}, nil
}

// Datasource is an example datasource which can respond to data queries, reports
// its health and has streaming skills.
type Datasource struct {
	Client     *fasthttp.Client
	host       string
	apiKey     string
	maxBuckets int32
}

type datasourceSettings struct {
	MaxBuckets int32  `json:"maxBuckets"`
	Url        string `json:"url"`
}

func (d *Datasource) solanaMetricsUrl(metricType MetricType, programId string, ixName string, start time.Time, end time.Time, bucket int32) string {
	uri := fmt.Sprintf("%s/query/solana/instructions/%s/%s?start=%s&end=%s&bucketSeconds=%d", d.host, programId, metricType, start.Format(time.RFC3339), end.Format(time.RFC3339), bucket)
	if ixName != "" {
		uri = fmt.Sprintf("%s&instructionName=%s", uri, ixName)
	}
	return uri
}

func (d *Datasource) Dispose() {
}

func (d *Datasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	// create response struct
	response := backend.NewQueryDataResponse()

	// loop over queries and execute them individually.
	for _, q := range req.Queries {
		res := d.query(ctx, req.PluginContext, q)

		// save the response in a hashmap
		// based on with RefID as identifier
		response.Responses[q.RefID] = res
	}

	return response, nil
}

type queryPayload struct {
	QueryType string `json:"queryType"`
	ProgramId string `json:"programId"`
	IxName    string `json:"instructionName"`
	TopN      int32  `json:"topN"`
}

type queryModel struct {
	Payload queryPayload `json:"payload"`
}

type invocationsBucket struct {
	Count  int32     `json:"count"`
	Status string    `json:"status"`
	IxName string    `json:"instructionName"`
	Time   time.Time `json:"time"`
}

type signersBucket struct {
	Time  time.Time `json:"time"`
	Count int32     `json:"count"`
}

type programEventBucket struct {
	Count     int32     `json:"count"`
	Time      time.Time `json:"time"`
	Authority string    `json:"authority"`
	Status    string    `json:"status"`
	Action    string    `json:"action"`
}

type invocationsMetricsResponse struct {
	Buckets []invocationsBucket `json:"buckets"`
}

type programEventsResponse struct {
	Buckets []programEventBucket `json:"buckets"`
}

type signersMetricsResponse struct {
	Buckets []signersBucket `json:"buckets"`
}

type instructionInvocationsResponse struct {
	times        []time.Time
	values       []int32
	statusLabels []string
	ixNameLabels []string
}

type invocationHeap []*invocationHeapItem

type invocationHeapItem struct {
	ixName string
	count  int32
	index  int
}

func (h invocationHeap) Len() int           { return len(h) }
func (h invocationHeap) Less(i, j int) bool { return h[i].count > h[j].count }
func (h invocationHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}

func (h *invocationHeap) Push(x any) {
	n := len(*h)
	item := x.(*invocationHeapItem)
	item.index = n
	*h = append(*h, item)
}

func (h *invocationHeap) Pop() any {
	old := *h
	n := len(old)
	item := old[n-1]
	old[n-1] = nil // avoid memory leak
	item.index = -1
	*h = old[0 : n-1]
	return item
}

func (pq *invocationHeap) update(item *invocationHeapItem, total int32) {
	item.count += total
	heap.Fix(pq, item.index)
}

func (h invocationHeap) topN(n int32) map[string]bool {
	topN := make(map[string]bool)
	if n > int32(len(h)) {
		n = int32(len(h))
	}
	var i int32
	for i = 0; i < n; i++ {
		item := heap.Pop(&h).(*invocationHeapItem)
		topN[item.ixName] = true
	}
	return topN
}

func getInstructionInvocations(resp *fasthttp.Response) (instructionInvocationsResponse, error) {
	var invocationsResponse invocationsMetricsResponse
	err := json.Unmarshal(resp.Body(), &invocationsResponse)
	if err != nil {
		return instructionInvocationsResponse{}, err
	}
	times := make([]time.Time, len(invocationsResponse.Buckets))
	values := make([]int32, len(invocationsResponse.Buckets))
	statusLabels := make([]string, len(invocationsResponse.Buckets))
	ixNameLabels := make([]string, len(invocationsResponse.Buckets))
	//Faster than append since we pre allocated
	//Revers iter for Long -> wide conversion
	for i, n := range invocationsResponse.Buckets {
		times[i] = n.Time
		values[i] = int32(n.Count)
		statusLabels[i] = n.Status
		ixNameLabels[i] = n.IxName
	}
	return instructionInvocationsResponse{
		times:        times,
		values:       values,
		statusLabels: statusLabels,
		ixNameLabels: ixNameLabels,
	}, nil
}

func (d *Datasource) query(_ context.Context, pCtx backend.PluginContext, query backend.DataQuery) backend.DataResponse {
	// Unmarshal the JSON into our queryModel.
	var qm queryModel
	log.DefaultLogger.Info("Query", "query", query)
	err := json.Unmarshal(query.JSON, &qm)
	if err != nil {
		return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("json unmarshal: %v", err.Error()))
	}
	log.DefaultLogger.Info("Query", "query", qm)
	req := fasthttp.AcquireRequest()
	req.Header.SetMethod("GET")
	req.Header.SetContentType("application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", d.apiKey))
	buckets := clampBucketSeconds(query.TimeRange.From, query.TimeRange.To, int32(query.Interval.Seconds()), d.maxBuckets)
	qt := getQueryType(MetricType(qm.Payload.QueryType))
	req.SetRequestURI(d.solanaMetricsUrl(qt, qm.Payload.ProgramId, qm.Payload.IxName, query.TimeRange.From, query.TimeRange.To, buckets))
	log.DefaultLogger.Info("URI", string(req.URI().FullURI()))
	resp := fasthttp.AcquireResponse()
	httpError := d.Client.Do(req, resp)
	fasthttp.ReleaseRequest(req)
	if httpError != nil {
		return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("upstream error: %v", httpError.Error()))
	}
	if resp.StatusCode() != 200 {
		return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("upstream error: %v", resp.StatusCode()))
	}
	var response backend.DataResponse
	frame := data.NewFrame("Long").SetMeta(&data.FrameMeta{
		Type:        data.FrameTypeTimeSeriesLong,
		TypeVersion: data.FrameTypeVersion{0, 1},
	})
	switch qm.Payload.QueryType {
	case string(Invocations):
		{
			r, err := getInstructionInvocations(resp)

			if err != nil {
				return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("json unmarshal: %v", err.Error()))
			}

			frame.Fields = append(frame.Fields,
				data.NewField("time", nil, r.times),
				data.NewField("count", nil, r.values),
				data.NewField("status", nil, r.statusLabels),
				data.NewField("instructionName", nil, r.ixNameLabels),
			)

		}
	case string(TopInstructions):
		{
			r, err := getInstructionInvocations(resp)
			log.DefaultLogger.Info("TopInstructions", "r", r)
			if err != nil {
				return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("json unmarshal: %v", err.Error()))
			}
			log.DefaultLogger.Info("TopInstructions", "topN", qm.Payload.TopN)
			var ixTotalsMap = invocationHeap{}
			var seen = make(map[string]*invocationHeapItem)
			heap.Init(&ixTotalsMap)
			for i := 0; i < len(r.times); i++ {
				if _, ok := seen[r.ixNameLabels[i]]; !ok {
					seen[r.ixNameLabels[i]] = &invocationHeapItem{
						count:  r.values[i],
						ixName: r.ixNameLabels[i],
					}
					heap.Push(&ixTotalsMap, seen[r.ixNameLabels[i]])
				} else {
					ixTotalsMap.update(seen[r.ixNameLabels[i]], r.values[i])
				}
			}
			log.DefaultLogger.Info("made it here")
			top := ixTotalsMap.topN(qm.Payload.TopN)
			lenb := len(r.times)
			times := make([]time.Time, lenb)
			values := make([]int32, lenb)
			statusLabels := make([]string, lenb)
			ixNameLabels := make([]string, lenb)

			for i := 0; i < len(r.times); i++ {
				ix_name := r.ixNameLabels[i]
				if _, ok := top[ix_name]; ok {

					times = append(times, r.times[i])
					values = append(values, r.values[i])
					statusLabels = append(statusLabels, r.statusLabels[i])
					ixNameLabels = append(ixNameLabels, r.ixNameLabels[i])
				}
			}

			frame.Fields = append(frame.Fields,
				data.NewField("time", nil, times),
				data.NewField("count", nil, values),
				data.NewField("status", nil, statusLabels),
				data.NewField("instructionName", nil, ixNameLabels),
			)

		}
	case string(UniqueSigners):
		{
			var signersResponse signersMetricsResponse
			err = json.Unmarshal(resp.Body(), &signersResponse)
			if err != nil {
				return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("json unmarshal: %v", err.Error()))
			}
			times := make([]time.Time, len(signersResponse.Buckets))
			values := make([]int32, len(signersResponse.Buckets))

			for _, bucket := range signersResponse.Buckets {
				times = append(times, bucket.Time)
				values = append(values, bucket.Count)
			}
			frame.Fields = append(frame.Fields,
				data.NewField("time", nil, times),
				data.NewField("count", nil, values),
			)
		}
	case string(Failures):
		{
			var invocationsResponse invocationsMetricsResponse
			err = json.Unmarshal(resp.Body(), &invocationsResponse)
			if err != nil {
				return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("json unmarshal: %v", err.Error()))
			}
			times := []time.Time{}
			values := []int32{}
			ixNameLabels := []string{}
			//Faster than append since we pre allocated
			//Revers iter for Long -> wide conversion
			for _, n := range invocationsResponse.Buckets {
				if n.Status != "success" {
					times = append(times, n.Time)
					values = append(values, int32(n.Count))
					ixNameLabels = append(ixNameLabels, n.IxName)
				}
			}

			frame.Fields = append(frame.Fields,
				data.NewField("time", nil, times),
				data.NewField("count", nil, values),
				data.NewField("instructionName", nil, ixNameLabels),
			)

			if err != nil {
				return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("long to wide: %v", err.Error()))
			}
		}
	case string(FailureRate):
		{
			var invocationsResponse invocationsMetricsResponse
			err = json.Unmarshal(resp.Body(), &invocationsResponse)
			if err != nil {
				return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("json unmarshal: %v", err.Error()))
			}
			times := make([]time.Time, len(invocationsResponse.Buckets))
			values := make([]float32, len(invocationsResponse.Buckets))
			var curr_time time.Time
			var curr_failure float32 = 0
			var curr_len float32 = 0
			for i, n := range invocationsResponse.Buckets[0 : len(invocationsResponse.Buckets)-1] {
				if i != 0 && n.Time != curr_time {
					curr_time = n.Time
					if n.Status != "success" {
						curr_failure++
					}
					curr_len++
				} else {
					times = append(times, curr_time)
					values = append(values, curr_failure/curr_len)
				}

			}
			frame.Fields = append(frame.Fields,
				data.NewField("time", nil, times),
				data.NewField("rate", nil, values),
			)
		}
	case string(FailedProgramDeployments):
		fallthrough
	case string(ProgramDeployments):

		{
			var programEvents programEventsResponse
			err = json.Unmarshal(resp.Body(), &programEvents)
			if err != nil {
				return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("json unmarshal: %v", err.Error()))
			}
			times := make([]time.Time, len(programEvents.Buckets))
			values := make([]int32, len(programEvents.Buckets))
			authorityLabels := make([]string, len(programEvents.Buckets))
			statusLabels := make([]string, len(programEvents.Buckets))
			actionLabels := make([]string, len(programEvents.Buckets))
			//Faster than append since we pre allocated
			//Revers iter for Long -> wide conversion
			log.DefaultLogger.Info("Program Events: ", programEvents.Buckets)
			for i, n := range programEvents.Buckets {
				times[i] = n.Time
				values[i] = int32(n.Count)
				statusLabels[i] = n.Status
				authorityLabels[i] = n.Authority
				actionLabels[i] = n.Action
			}

			frame.Fields = append(frame.Fields,
				data.NewField("time", nil, times),
				data.NewField("count", nil, values),
				data.NewField("authority", nil, statusLabels),
				data.NewField("status", nil, statusLabels),
				data.NewField("action", nil, actionLabels),
			)

			if err != nil {
				return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("long to wide: %v", err.Error()))
			}
		}
	default:
		return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("invalid query type: %s", qm.Payload.QueryType))
	}
	fasthttp.ReleaseResponse(resp)
	response.Frames = append(response.Frames, frame)
	return response
}

// CheckHealth handles health checks sent from Grafana to the plugin.
// The main use case for these health checks is the test button on the
// datasource configuration page which allows users to verify that
// a datasource is working as expected.
func (d *Datasource) CheckHealth(_ context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	var status = backend.HealthStatusOk
	var message = "Data source is working"

	return &backend.CheckHealthResult{
		Status:  status,
		Message: message,
	}, nil
}
