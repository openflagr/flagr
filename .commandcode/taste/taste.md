# Taste (C ntinuously Learned by [CommandCode][cmd])

# notifications
- In notification/event systems, the `operation` field must accurately reflect what happened to the specific component (use `OperationCreate` for creates, `OperationDelete` for deletes, `OperationUpdate` for actual updates), not use a generic "update" for all sub-resource mutations. Confidence: 0.85
