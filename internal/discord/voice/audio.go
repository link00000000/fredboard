package voice

import (
	"accidentallycoded.com/fredboard/v3/internal/audio/graph"
)

// TODO: This will require composite nodes that can contain a source and decoder node in one
// DiscordAudioGraph always takes the form:
// Source1 --------> Mixer -> Transformer -> ... -> Transformer -> DiscordSink
// Source2 -> Mixer --/
// Source3 --/
// ...
type DiscordAudioGraph struct {
	*graph.AudioGraph
}

func (graph *DiscordAudioGraph) AddSource(node *graph.AudioGraphNode) {

}
