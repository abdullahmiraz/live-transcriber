export type MeetingPhase = 'loading' | 'lobby' | 'joining' | 'in-call' | 'error';

export type JoinMediaMode = 'both' | 'audio' | 'video' | 'none';

export type MeetingTab = 'chat' | 'captions' | 'people';

export type VideoTile = {
	id: string;
	name: string;
	stream: MediaStream;
};

export type Caption = {
	key: string;
	participantId: string;
	name: string;
	original: string;
	translated: string;
};
