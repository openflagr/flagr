package mcp

// registerTools registers all Flagr MCP tools.
func (s *Server) registerTools() {
	s.registerFlagTools()
	s.registerSegmentTools()
	s.registerVariantTools()
	s.registerConstraintTools()
	s.registerTagTools()
	s.registerEvaluateTools()
}
