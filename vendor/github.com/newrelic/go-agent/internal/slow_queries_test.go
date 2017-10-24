package internal

import (
	"math/rand"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestEmptySlowQueriesData(t *testing.T) {
	slows := newSlowQueries(maxHarvestSlowSQLs)
	js, err := slows.Data("agentRunID", time.Now())
	if nil != js || nil != err {
		t.Error(string(js), err)
	}
}

func TestSlowQueriesBasic(t *testing.T) {
	txnSlows := newSlowQueries(maxTxnSlowQueries)
	txnSlows.observeInstance(slowQueryInstance{
		Duration:           2 * time.Second,
		DatastoreMetric:    "Datastore/statement/MySQL/users/INSERT",
		ParameterizedQuery: "INSERT INTO users (name, age) VALUES ($1, $2)",
		Host:               "db-server-1",
		PortPathOrID:       "3306",
		DatabaseName:       "production",
		StackTrace:         nil,
		QueryParameters: vetQueryParameters(map[string]interface{}{
			strings.Repeat("X", attributeKeyLengthLimit+1): "invalid-key",
			"invalid-value":                                struct{}{},
			"valid":                                        123,
		}),
	})
	harvestSlows := newSlowQueries(maxHarvestSlowSQLs)
	harvestSlows.Merge(txnSlows, "WebTransaction/Go/hello", "/zip/zap")
	js, err := harvestSlows.Data("agentRunID", time.Now())
	expect := CompactJSONString(`[[
	[
		"WebTransaction/Go/hello",
		"/zip/zap",
		3722056893,
		"INSERT INTO users (name, age) VALUES ($1, $2)",
		"Datastore/statement/MySQL/users/INSERT",
		1,
		2000,
		2000,
		2000,
		{
			"host":"db-server-1",
			"port_path_or_id":"3306",
			"database_name":"production",
			"query_parameters":{
				"valid":123
			}
		}
	]
]]`)
	if nil != err {
		t.Error(err)
	}
	if string(js) != expect {
		t.Error(string(js), expect)
	}
}

func TestSlowQueriesAggregation(t *testing.T) {
	max := 50
	slows := make([]slowQueryInstance, 3*max)
	for i := 0; i < max; i++ {
		num := i + 1
		str := strconv.Itoa(num)
		duration := time.Duration(num) * time.Second
		slow := slowQueryInstance{
			DatastoreMetric:    "Datastore/" + str,
			ParameterizedQuery: str,
		}
		slow.Duration = duration
		slow.TxnName = "Txn/0" + str
		slow.TxnURL = "/0" + str
		slows[i*3+0] = slow
		slow.Duration = duration + (100 * time.Second)
		slow.TxnName = "Txn/1" + str
		slow.TxnURL = "/1" + str
		slows[i*3+1] = slow
		slow.Duration = duration + (200 * time.Second)
		slow.TxnName = "Txn/2" + str
		slow.TxnURL = "/2" + str
		slows[i*3+2] = slow
	}
	sq := newSlowQueries(10)
	seed := int64(99) // arbitrary fixed seed
	r := rand.New(rand.NewSource(seed))
	perm := r.Perm(max * 3)
	for _, idx := range perm {
		sq.observeInstance(slows[idx])
	}
	js, err := sq.Data("agentRunID", time.Now())
	expect := CompactJSONString(`[[
	["Txn/241","/241",2296612630,"41","Datastore/41",1,241000,241000,241000,{}],
	["Txn/242","/242",2279835011,"42","Datastore/42",2,384000,142000,242000,{}],
	["Txn/243","/243",2263057392,"43","Datastore/43",2,386000,143000,243000,{}],
	["Txn/244","/244",2380500725,"44","Datastore/44",3,432000,44000,244000,{}],
	["Txn/247","/247",2330167868,"47","Datastore/47",2,394000,147000,247000,{}],
	["Txn/245","/245",2363723106,"45","Datastore/45",2,290000,45000,245000,{}],
	["Txn/250","/250",2212577440,"50","Datastore/50",1,250000,250000,250000,{}],
	["Txn/246","/246",2346945487,"46","Datastore/46",2,392000,146000,246000,{}],
	["Txn/249","/249",2430833582,"49","Datastore/49",3,447000,49000,249000,{}],
	["Txn/248","/248",2447611201,"48","Datastore/48",3,444000,48000,248000,{}]
]]`)
	if nil != err {
		t.Error(err)
	}
	if string(js) != expect {
		t.Error(string(js), expect)
	}
}
