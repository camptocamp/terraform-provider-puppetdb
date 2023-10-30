package log

func NodeFields(certificateName string) map[string]any {
	return map[string]any{
		"certificate_name": certificateName,
	}
}
