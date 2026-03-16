import { create } from 'zustand';
import type { ChatStoreState } from './chat/chatStore.types';
import { createConnectionSlice } from './chat/connectionSlice';
import { createUiSlice } from './chat/uiSlice';
import { createRoomSlice } from './chat/roomSlice';
import { createDmSlice } from './chat/dmSlice';

export const useChatStore = create<ChatStoreState>((set, get, store) => ({
    ...createConnectionSlice(set, get, store),
    ...createUiSlice(set, get, store),
    ...createRoomSlice(set, get, store),
    ...createDmSlice(set, get, store),
}));

export * from './chat/chatStore.types';
export type { ReactionSummary } from './chat/chatStore.types';
