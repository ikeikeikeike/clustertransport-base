package clustertransport

// ElasticsearchMethod is
type ElasticsearchMethod struct{}

// CarryOut is
func (m *ElasticsearchMethod) CarryOut() []string {
	return []string{"1", "2"}
}
