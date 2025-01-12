package chatgpt

import "github.com/sashabaranov/go-openai"

type Queue struct {
	data []openai.ChatCompletionMessage
	size int
}

func NewQueue(size int) *Queue {
	return &Queue{
		data: make([]openai.ChatCompletionMessage, 0, size*2+1),
		size: size * 2,
	}
}

func (q *Queue) Enqueue(msg openai.ChatCompletionMessage) {
	q.data = append(q.data, msg)
	if len(q.data) > q.size {
		q.data = q.data[2:]
	}
}

func (q *Queue) Dequeue() openai.ChatCompletionMessage {
	if len(q.data) == 0 {
		return openai.ChatCompletionMessage{}
	}
	msg := q.data[0]
	q.data = q.data[1:]
	return msg
}

func (q *Queue) Size() int {
	return len(q.data)
}
func (q *Queue) IsEmpty() bool {
	return len(q.data) == 0
}

func (q *Queue) Get() []openai.ChatCompletionMessage {
	return q.data
}
