#!/bin/bash

# shellcheck disable=SC1091
source /vendor/shakedown/shakedown.sh

uuid_str() {
    tr </dev/urandom -dc '[:lower:]' | fold -w 16 | head -n 1
}

step_1_test_health() {
    flagr_url=$1:18000/api/v1

    shakedown GET "$flagr_url"/health
    status 200
    content_type 'application/json'
}

step_2_test_crud_flag() {
    flagr_url=$1:18000/api/v1

    description_1=$(uuid_str)
    description_2=$(uuid_str)
    description_3=$(uuid_str)
    key_1=$(uuid_str)
    key_2=$(uuid_str)
    key_3=$(uuid_str)

    ################################################
    # Test create flag
    ################################################
    shakedown POST "$flagr_url"/flags -H 'Content-Type:application/json' -d "{\"description\": \"$description_1\", \"key\": \"$key_1\"}"
    status 200
    contains "$description_1"
    contains "$key_1"

    shakedown POST "$flagr_url"/flags -H 'Content-Type:application/json' -d "{\"description\": \"$description_2\"}"
    status 200
    contains "$description_2"
    matches '"key":"[a-z0-9]+"'

    shakedown GET "$flagr_url"/flags -H 'Content-Type:application/json'
    status 200
    contains "$description_1"
    contains "$key_1"

    ################################################
    # Test put flag
    ################################################
    shakedown PUT "$flagr_url"/flags/1 -H 'Content-Type:application/json' -d "{\"description\": \"$description_3\", \"key\": \"$key_3\", \"enabled\": true}"
    status 200
    contains "$description_3"
    contains "$key_3"

    shakedown GET "$flagr_url"/flags/1 -H 'Content-Type:application/json'
    status 200
    contains "$description_3"
    contains "$key_3"

    shakedown PUT "$flagr_url"/flags/1/enabled -H 'Content-Type:application/json' -d "{\"enabled\": false}"
    status 200
    matches '"enabled":false'

    shakedown PUT "$flagr_url"/flags/1/enabled -H 'Content-Type:application/json' -d "{\"enabled\": true}"
    status 200
    matches '"enabled":true'

    ################################################
    # Test put flag entity type
    ################################################
    shakedown PUT "$flagr_url"/flags/1 -H 'Content-Type:application/json' -d "{\"dataRecordsEnabled\": true, \"entityType\": \"candidate_resource_id\"}"
    status 200
    matches '"dataRecordsEnabled":true'
    matches '"entityType":"candidate_resource_id"'

    shakedown GET "$flagr_url"/flags/entity_types -H 'Content-Type:application/json'
    status 200
    contains 'candidate_resource_id'

    ################################################
    # Test query flags
    ################################################
    shakedown PUT "$flagr_url"/flags/1 -H 'Content-Type:application/json' -d "{\"description\": \"flag_1\", \"key\": \"key_1\"}"
    status 200

    shakedown PUT "$flagr_url"/flags/2 -H 'Content-Type:application/json' -d "{\"description\": \"flag_2\", \"key\": \"key_2\"}"
    status 200

    shakedown GET "$flagr_url"/flags?description=flag_1 -H 'Content-Type:application/json'
    status 200
    matches '"description":"flag_1"'
    matches '"key":"key_1"'

    shakedown GET "$flagr_url"/flags?description=flag_1 -H 'Content-Type:application/json'
    status 200
    matches '"description":"flag_1"'
    matches '"key":"key_1"'

    shakedown GET "$flagr_url/flags?description_like=flag_" -H 'Content-Type:application/json'
    status 200
    matches '"description":"flag_1"'
    matches '"description":"flag_2"'

    shakedown GET "$flagr_url/flags?description_like=flag_&limit=1" -H 'Content-Type:application/json'
    status 200
    matches '"description":"flag_1"'
    matches '"description":"flag_[^2]"'

    shakedown GET "$flagr_url/flags?description_like=flag_&limit=1&offset=1" -H 'Content-Type:application/json'
    status 200
    matches '"description":"flag_[^1]"'
    matches '"description":"flag_2"'

    shakedown GET "$flagr_url/flags?key=key_1" -H 'Content-Type:application/json'
    status 200
    matches '"description":"flag_1"'
    matches '"key":"key_1"'
}

step_3_test_crud_segment() {
    flagr_url=$1:18000/api/v1

    description_1=$(uuid_str)
    description_2=$(uuid_str)
    description_3=$(uuid_str)

    ################################################
    # Test create segment
    ################################################
    shakedown POST "$flagr_url"/flags/1/segments -H 'Content-Type:application/json' -d "{\"description\": \"$description_1\", \"rolloutPercent\": 100}"
    status 200
    matches "\"description\":\"$description_1\""
    matches "\"rolloutPercent\":100"

    shakedown POST "$flagr_url"/flags/1/segments -H 'Content-Type:application/json' -d "{\"description\": \"$description_2\", \"rolloutPercent\": 0}"
    status 200
    matches "\"description\":\"$description_2\""
    matches "\"rolloutPercent\":0"

    shakedown GET "$flagr_url"/flags/1/segments -H 'Content-Type:application/json'
    status 200
    matches "\"description\":\"$description_1\""
    matches "\"description\":\"$description_2\""
    matches "\"rolloutPercent\":100"
    matches "\"rolloutPercent\":0"
    matches "\"rank\":999"

    ################################################
    # Test put segment
    ################################################
    shakedown PUT "$flagr_url"/flags/1/segments/1 -H 'Content-Type:application/json' -d "{\"description\": \"$description_3\", \"rolloutPercent\": 100}"
    status 200
    matches "\"description\":\"$description_3\""

    shakedown PUT "$flagr_url"/flags/1/segments/2 -H 'Content-Type:application/json' -d "{\"description\": \"$description_2\", \"rolloutPercent\": 100}"
    status 200
    matches "\"description\":\"$description_2\""
    matches "\"rolloutPercent\":100"

    ################################################
    # Test reorder segment
    ################################################
    shakedown PUT "$flagr_url"/flags/1/segments/reorder -H 'Content-Type:application/json' -d "{\"segmentIDs\": [2, 1]}"
    status 200

    shakedown GET "$flagr_url"/flags/1/segments -H 'Content-Type:application/json'
    status 200
    matches "\"rank\":0"
    matches "\"rank\":1"

    shakedown PUT "$flagr_url"/flags/1/segments/reorder -H 'Content-Type:application/json' -d "{\"segmentIDs\": [1, 2]}"
    status 200
}

step_4_test_crud_constraint() {
    flagr_url=$1:18000/api/v1

    ################################################
    # Test create constraint
    ################################################
    shakedown POST "$flagr_url"/flags/1/segments/1/constraints -H 'Content-Type:application/json' -d '{"property": "property_1","operator": "EQ","value": "\"value_1\""}'
    status 200
    matches "\"property\":\"property_1\""
    contains "value_1"

    ################################################
    # Test put constraint
    ################################################
    shakedown PUT "$flagr_url"/flags/1/segments/1/constraints/1 -H 'Content-Type:application/json' -d '{"property": "property_1","operator": "EQ","value": "\"value_2\""}'
    status 200
    matches "\"property\":\"property_1\""
    contains "value_2"
}

step_5_test_crud_variant() {
    flagr_url=$1:18000/api/v1

    key_1=$(uuid_str)
    key_2=$(uuid_str)

    ################################################
    # Test create variant
    ################################################
    shakedown POST "$flagr_url"/flags/1/variants -H 'Content-Type:application/json' -d "{\"key\": \"$key_1\"}"
    status 200
    matches "\"key\":\"$key_1\""

    shakedown POST "$flagr_url"/flags/1/variants -H 'Content-Type:application/json' -d "{\"key\": \"$key_2\"}"
    status 200
    matches "\"key\":\"$key_2\""

    shakedown GET "$flagr_url"/flags/1/variants -H 'Content-Type:application/json'
    status 200
    matches "\"key\":\"$key_1\""
    matches "\"key\":\"$key_2\""

    ################################################
    # Test put variant
    ################################################
    shakedown PUT "$flagr_url"/flags/1/variants/1 -H 'Content-Type:application/json' -d "{\"key\": \"key_1\"}"
    status 200
    matches "\"key\":\"key_1\""

    shakedown PUT "$flagr_url"/flags/1/variants/2 -H 'Content-Type:application/json' -d "{\"key\": \"key_2\"}"
    status 200
    matches "\"key\":\"key_2\""
}

step_6_test_crud_distribution() {
    flagr_url=$1:18000/api/v1

    ################################################
    # Test put distribution
    ################################################
    shakedown PUT "$flagr_url"/flags/1/segments/1/distributions -H 'Content-Type:application/json' -d '{"distributions": [{"percent": 100, "variantKey": "key_1", "variantID": 1}]}'
    status 200
    matches "\"percent\":100"
    matches "\"variantKey\":\"key_1\""
    matches "\"variantID\":1"

    shakedown GET "$flagr_url"/flags/1/segments/1/distributions -H 'Content-Type:application/json'
    status 200
    matches "\"percent\":100"
    matches "\"variantKey\":\"key_1\""
    matches "\"variantID\":1"

    shakedown PUT "$flagr_url"/flags/1/segments/2/distributions -H 'Content-Type:application/json' -d '{"distributions": [{"percent": 100, "variantKey": "key_2", "variantID": 2}]}'
    status 200
    matches "\"percent\":100"
    matches "\"variantKey\":\"key_2\""
    matches "\"variantID\":2"

    shakedown GET "$flagr_url"/flags/1/segments/2/distributions -H 'Content-Type:application/json'
    status 200
    matches "\"percent\":100"
    matches "\"variantKey\":\"key_2\""
    matches "\"variantID\":2"
}

step_7_test_evaluation() {
    flagr_url=$1:18000/api/v1
    sleep 5

    ################################################
    # Test post evaluation
    ################################################
    shakedown POST "$flagr_url"/evaluation -H 'Content-Type:application/json' -d '{"entityID": "abc1234", "entityType": "candidate", "entityContext": {"property_1": "value_2"}, "flagID": 1}'
    status 200
    matches "\"variantKey\":\"key_1\""
    matches "\"variantID\":1"
    matches "\"flagID\":1"
    matches "\"segmentID\":1"

    shakedown POST "$flagr_url"/evaluation -H 'Content-Type:application/json' -d '{"entityID": "abc1234", "entityType": "candidate", "flagID": 1}'
    status 200
    matches "\"variantKey\":\"key_2\""
    matches "\"variantID\":2"
    matches "\"flagID\":1"
    matches "\"segmentID\":2"
}

step_8_test_preload() {
    flagr_url=$1:18000/api/v1

    ################################################
    # Test preload for /flags (depends on ?preload=true/false
    ################################################
    shakedown GET "$flagr_url"/flags
    status 200
    matches "\"variants\":\[\]"
    matches "\"segments\":\[\]"

    shakedown GET "$flagr_url"/flags?preload=true
    status 200
    matches "\"variantKey\":\"key_1\""
    matches "\"variantID\":1"

    ################################################
    # Test preload for /flag
    # always preload for getting a single flag
    ################################################
    shakedown GET "$flagr_url"/flags/1
    status 200
    matches "\"variantKey\":\"key_1\""
    matches "\"variantID\":1"
}

step_9_test_export() {
    flagr_url=$1:18000/api/v1

    ################################################
    # Test export
    ################################################
    shakedown GET "$flagr_url"/export/sqlite
    status 200

    shakedown GET "$flagr_url"/export/eval_cache/json
    status 200
    matches "\"VariantKey\":\"key_1\""
    matches "\"VariantID\":1"
}

step_10_test_crud_tag() {
    flagr_url=$1:18000/api/v1
    # sleep 5

    ################################################
    # Test create tags
    ################################################
    shakedown POST "$flagr_url"/flags/1/tags -H 'Content-Type:application/json' -d "{\"value\": \"value_1\"}"
    status 200
    matches "\"value\":\"value_1\""

    shakedown POST "$flagr_url"/flags/1/tags -H 'Content-Type:application/json' -d "{\"value\": \"value_2\"}"
    status 200
    matches "\"value\":\"value_2\""

    shakedown GET "$flagr_url"/flags/1/tags -H 'Content-Type:application/json'
    status 200
    matches "\"value\":\"value_1\""
    matches "\"value\":\"value_2\""
}

step_11_test_tag_batch_evaluation() {
    flagr_url=$1:18000/api/v1
    sleep 5

    shakedown POST "$flagr_url"/evaluation/batch -H 'Content-Type:application/json' -d '{"entities":[{ "entityType": "externalalert", "entityContext": {"property_1": "value_2"} }],"flagTags": ["value_1"], "enableDebug": false }'
    status 200
    matches "\"flagID\":1"
    matches "\"variantKey\":\"key_1\""
    matches "\"variantID\":1"

}

step_12_test_tag_operator_batch_evaluation() {
    flagr_url=$1:18000/api/v1
    sleep 5

    shakedown POST "$flagr_url"/evaluation/batch -H 'Content-Type:application/json' -d '{"entities":[{ "entityType": "externalalert", "entityContext": {"property_1": "value_2"} }],"flagTags": ["value_1", "value_2"], "flagTagsOperator": "ALL", "enableDebug": false }'
    status 200
    matches "\"flagID\":1"
    matches "\"variantKey\":\"key_1\""
    matches "\"variantID\":1"

    shakedown POST "$flagr_url"/evaluation/batch -H 'Content-Type:application/json' -d '{"entities":[{ "entityType": "externalalert", "entityContext": {"property_1": "value_2"} }],"flagTags": ["value_1", "value_3"], "flagTagsOperator": "ANY", "enableDebug": false }'
    status 200
    matches "\"flagID\":1"
    matches "\"variantKey\":\"key_1\""
    matches "\"variantID\":1"

}

start_test() {
    flagr_host=$1
    echo -e "\e[32m                \e[0m"
    echo -e "\e[32m===========================================\e[0m"
    echo -e "\e[32mStart testing $1\e[0m"
    echo -e "\e[32m===========================================\e[0m"

    /vendor/wait-for-it/wait-for-it.sh "$flagr_host:18000" -t 30

    step_1_test_health "$flagr_host"
    step_2_test_crud_flag "$flagr_host"
    step_3_test_crud_segment "$flagr_host"
    step_4_test_crud_constraint "$flagr_host"
    step_5_test_crud_variant "$flagr_host"
    step_6_test_crud_distribution "$flagr_host"
    step_7_test_evaluation "$flagr_host"
    step_8_test_preload "$flagr_host"
    step_9_test_export "$flagr_host"
    step_10_test_crud_tag "$flagr_host"
    step_11_test_tag_batch_evaluation "$flagr_host"
    step_12_test_tag_operator_batch_evaluation "$flagr_host"
}

start() {
    start_test flagr_with_sqlite
    start_test flagr_with_mysql
    start_test flagr_with_mysql8
    start_test flagr_with_postgres9
    start_test flagr_with_postgres13

    # for backward compatibility with checkr/flagr
    start_test checkr_flagr_with_sqlite
}

start
