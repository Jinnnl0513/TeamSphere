export interface User {
    id: number;
    username: string;
    avatar_url: string;
    bio?: string;
    profile_color?: string;
    role: string;
}

export interface Room {
    id: number;
    name: string;
    description: string;
    creator_id: number;
    created_at: string;
}

export interface Message {
    id: number;
    content: string;
    user_id: number;
    room_id?: number;
    msg_type: string;
    client_msg_id?: string;
    user?: User;
    mentions?: number[];
    deleted_at?: string;
    created_at: string;
}

export interface FriendRequest {
    id: number;
    user_id: number;
    friend_id: number;
    status: string;
    created_at: string;
    user?: User; // Depending on if it's incoming or outgoing, typically the "other" user
}
