package jetlib

func TopicForUp(topic_base string) string {
	return "u." + topic_base
}

func TopicForDown(topic_base string) string {
	return "d." + topic_base
}

func TopicForJet(topic_base string) string {
	return "j_" + topic_base
}

func TopicForBase(topic string) string {
	return topic[2:]
}
