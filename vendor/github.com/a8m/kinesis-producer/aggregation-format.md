# KPL Aggregated Record Format
> Note: This file taken from: [amazon-kinesis-producer](https://github.com/awslabs/amazon-kinesis-producer/blob/master/aggregation-format.md)

## Intro 

The Amazon Kinesis Producer Library (KPL) aggregates multiple logical user records into a single Amazon Kinesis record for efficient puts.

We use Google protocol buffers (protobuf) to create a binary file format for this. The Amazon Kinesis Client Library (KCL) implements deaggregation based on this format on the consumer side. 

This document contains the format used. Developers may use this information to produce aggregated records from their own code that will be compatible with the KCL's deaggregation logic.

## Format

All of the user data is contained in a protobuf message. To this, we add a magic number and a checksum. The overall format is as follows:

```
0               4                  N          N+15
+---+---+---+---+==================+---+...+---+
|  MAGIC NUMBER | PROTOBUF MESSAGE |    MD5    |
+---+---+---+---+==================+---+...+---+

```

The magic number contains the 4 bytes `0xF3 0x89 0x9A 0xC2`.

The protobuf message is as follows:

```
message AggregatedRecord {
  repeated string partition_key_table     = 1;
  repeated string explicit_hash_key_table = 2;
  repeated Record records                 = 3;
}
```

The sub-messages are as follows:

```
message Tag {
  required string key   = 1;
  optional string value = 2;
}

message Record {
  required uint64 partition_key_index     = 1;
  optional uint64 explicit_hash_key_index = 2;
  required bytes  data                    = 3;
  repeated Tag    tags                    = 4;
}
```

Note: we use the proto2 language (not proto3).

The protobuf message allows more efficient partition and explicit hash key packing by allowing multiple records to point to the same key in a table. This feature is optional; implementations can simply store the keys of every record as a separate entry in the tables, even if two or more of them are the same.

The key tables are zero-indexed; they are simply arrays, and the key indices are indices into those arrays.

Tags are not yet implemented in the KPL and KCL APIs.

Lastly, the 16-byte MD5 checksum is computed over the bytes of the serialized protobuf message.
