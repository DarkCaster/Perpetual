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
	MultilineTaggedFragment
)

type Fragment struct {
	Type         FragmentType
	Contents     string
	FileName     string
	FileNameTags string
}

type Message struct {
	Type      MessageType
	Fragments []Fragment
	RawText   string
}

func NewMessage(messageType MessageType) Message {
	return Message{Type: messageType, Fragments: []Fragment{}, RawText: ""}
}

func addMessageFragment(message Message, fragmentType FragmentType, contents string, filename string, tags string) Message {
	message.Fragments = append(message.Fragments, Fragment{Type: fragmentType, Contents: contents, FileName: filename, FileNameTags: tags})
	return message
}

func SetRawResponse(message Message, rawResponse string) Message {
	if message.Type != RealAIResponse && message.Type != SimulatedAIResponse {
		panic("SetRawResponse only can be issued on RealAIResponse or SimulatedAIResponse messages")
	}
	message.RawText = rawResponse
	return message
}

func AddPlainTextFragment(message Message, text string) Message {
	return addMessageFragment(message, PlainTextFragment, text, "", "")
}

func AddIndexFragment(message Message, filename string, tags []string) Message {
	bTags, err := json.Marshal(tags)
	if err != nil {
		panic(err)
	}
	return addMessageFragment(message, IndexFragment, "", filename, string(bTags))
}

func AddTaggedFragment(message Message, contents string, tags []string) Message {
	bTags, err := json.Marshal(tags)
	if err != nil {
		panic(err)
	}
	return addMessageFragment(message, TaggedFragment, contents, "", string(bTags))
}

func AddMultilineTaggedFragment(message Message, multilineContent string, tags []string) Message {
	bTags, err := json.Marshal(tags)
	if err != nil {
		panic(err)
	}
	return addMessageFragment(message, MultilineTaggedFragment, multilineContent, "", string(bTags))
}

func AddFileFragment(message Message, filename string, contents string, tags []string) Message {
	bTags, err := json.Marshal(tags)
	if err != nil {
		panic(err)
	}
	return addMessageFragment(message, FileFragment, contents, filename, string(bTags))
}
