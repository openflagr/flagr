import httplib
import random
import json
import datetime

conn = httplib.HTTPConnection("localhost:18000")
conn_elastic = httplib.HTTPConnection("localhost:9200")
headers = { 'content-type': "application/json" }

def random_payload():
    entity_id = random.randint(1, 1000000)
    d = {
            "entityID": str(entity_id),
            "entityType": "user",
            "entityContext": {
                "state": random.choice(['CA', 'NY', 'VA']),
                "dl_state": random.choice(['CA', 'NY', 'VA']),
            },
            "flagID": 2,
            "enableDebug": True
        }
    return json.dumps(d)

def index_elastic(payload):
    conn_elastic.request("POST", "/flagr/flagr-records", payload, headers)
    res = conn_elastic.getresponse()
    data = res.read()
    print(data.decode("utf-8"))

while 1:
    t = datetime.datetime.now()
    conn.request("POST", "/api/v1/evaluation", random_payload(), headers)
    res = conn.getresponse()
    print(str((datetime.datetime.now() - t).total_seconds() * 1000) + 'ms')

    data = res.read()
    index_elastic(data)
    print(data.decode("utf-8"))
