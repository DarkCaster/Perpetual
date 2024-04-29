package llm

import "encoding/json"

type MessageType int

const (
	UserRequest MessageType = iota
	SimulatedAIResponse
	RealAIResponse
)

type FragmentType int

const (
	PlainTextFragment FragmentType = iota
	IndexFragment
	FileFragment
	TaggedFragment
)

type Fragment struct {
	Type     FragmentType
	Payload  string
	Metadata string
}

type Message struct {
	IsLogged  bool //TODO: make it pointer to bool, so all copies will share the same status ?
	Type      MessageType
	Fragments []Fragment
	RawText   string
}

func NewMessage(messageType MessageType) Message {
	return Message{IsLogged: false, Type: messageType, Fragments: []Fragment{}, RawText: ""}
}

func addMessageFragment(message Message, fragmentType FragmentType, payload string, metadata string) Message {
	message.Fragments = append(message.Fragments, Fragment{Type: fragmentType, Payload: payload, Metadata: metadata})
	return message
}

func SetRawResponse(message Message, rawResponse string) Message {
	if message.Type != RealAIResponse {
		panic("SetRawResponse only can be issued on RealAIResponse messages")
	}
	message.RawText = rawResponse
	return message
}

func AddPlainTextFragment(message Message, text string) Message {
	return addMessageFragment(message, PlainTextFragment, text, "")
}

func AddIndexFragment(message Message, index string) Message {
	return addMessageFragment(message, IndexFragment, index, "")
}

func AddTaggedFragment(message Message, payload string, tags []string) Message {
	bTags, err := json.Marshal(tags)
	if err != nil {
		panic(err)
	}
	return addMessageFragment(message, TaggedFragment, payload, string(bTags))
}

func AddFileFragment(message Message, filename string, contents string) Message {
	return addMessageFragment(message, FileFragment, contents, filename)
}
