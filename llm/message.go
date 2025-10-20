package llm

import (
	"github.com/DarkCaster/Perpetual/utils"
)

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
	FileNameTags utils.TagPair
}

type Message struct {
	Type      MessageType
	Fragments []Fragment
	RawText   string
}

func NewMessage(messageType MessageType) Message {
	return Message{Type: messageType, Fragments: []Fragment{}, RawText: ""}
}

func addMessageFragment(message Message, fragmentType FragmentType, contents string, filename string, tags utils.TagPair) Message {
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
	return addMessageFragment(message, PlainTextFragment, text, "", utils.TagPair{})
}

func AddIndexFragment(message Message, filename string, tags utils.TagPair) Message {
	return addMessageFragment(message, IndexFragment, "", filename, tags)
}

func AddTaggedFragment(message Message, contents string, tags utils.TagPair) Message {
	return addMessageFragment(message, TaggedFragment, contents, "", tags)
}

func AddMultilineTaggedFragment(message Message, multilineContent string, tags utils.TagPair) Message {
	return addMessageFragment(message, MultilineTaggedFragment, multilineContent, "", tags)
}

func AddFileFragment(message Message, filename string, contents string, tags utils.TagPair) Message {
	return addMessageFragment(message, FileFragment, contents, filename, tags)
}
