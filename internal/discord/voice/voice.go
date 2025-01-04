package voice

// Map of guild id to [VoiceSession]s
var voiceSessions map[string]VoiceSession = make(map[string]VoiceSession)

type VoiceSession struct{}
