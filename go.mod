module github.com/link00000000/fredboard/v3

go 1.23.6

require (
	github.com/bwmarrin/discordgo v0.28.1 // direct
	github.com/google/uuid v1.6.0 // direct
	golang.org/x/term v0.26.0 // direct
	layeh.com/gopus v0.0.0-20210501142526-1ee02d434e32 // direct
)

require (
	github.com/AllenDang/cimgui-go v1.3.0 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/link00000000/gaps v0.0.0-00010101000000-000000000000 // indirect
	github.com/link00000000/telemetry v0.0.0-00010101000000-000000000000 // indirect
	golang.org/x/crypto v0.0.0-20210421170649-83a5a9bb288b // indirect
	golang.org/x/sys v0.27.0 // indirect
)

replace (
	github.com/link00000000/gaps => ./pkg/gaps
	github.com/link00000000/go-telemetry => ./pkg/go-telemetry
)
