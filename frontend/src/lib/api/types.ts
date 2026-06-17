export type MeetingStatus = 'active' | 'ended';

export interface Meeting {
	id: string;
	slug: string;
	title: string;
	host_name: string;
	status: MeetingStatus | string;
	join_url: string;
	created_at: string;
	ended_at?: string;
}

export interface ChatMessage {
	id: string;
	meetingId: string;
	senderId: string;
	senderName: string;
	content: string;
	createdAt: string;
}
