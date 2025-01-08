package voice

// Map of guild id to [VoiceSession]s
var voiceSessions map[string]*VoiceSession = make(map[string]*VoiceSession)

type VoiceSession struct {
	guildId string

	OnSourceStoppedDelegate EventEmitter[func(Source)]
}

func (s *VoiceSession) AddEventListener() {
}

func NewVoiceSession(guildId string) *VoiceSession {
	return &VoiceSession{guildId: guildId}
}

func FindOrCreateSession(guildId string) *VoiceSession {
	if session, ok := voiceSessions[guildId]; ok {
		return session
	}

	return NewVoiceSession(guildId)
}

func DestroySession(guildId string) {
	if _, ok := voiceSessions[guildId]; ok {
		delete(voiceSessions, guildId)
	}
}
