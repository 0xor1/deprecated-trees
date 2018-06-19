package sortdir

func String(sortAsc bool) string {
	if sortAsc {
		return "asc"
	}
	return "desc"
}

func GtLtSymbol(sortAsc bool) string {
	if sortAsc {
		return ">"
	} else {
		return "<"
	}
}
