import { afterEach, describe, expect, it, vi } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import type { ChatMessage } from '../../stores/chatStore';
import { useScrollBehavior } from '../../pages/Chat/components/MessageList/hooks/useScrollBehavior';

function createMessage(id: number, created_at: string): ChatMessage {
  return {
    id,
    client_msg_id: String(id),
    room_id: 1,
    dm_id: null,
    user_id: 1,
    content: 'hi',
    msg_type: 'text',
    created_at,
    updated_at: created_at,
    deleted_at: undefined,
    user: {
      id: 1,
      username: 'alice',
      avatar_url: '',
      role: 'member',
      created_at,
      updated_at: created_at,
      deleted_at: undefined,
    },
    reactions: [],
    reply_to: undefined,
  } as ChatMessage;
}

function setupDom(scrollTop = 0, scrollHeight = 1000, clientHeight = 400) {
  const el = document.createElement('div');
  Object.defineProperty(el, 'scrollTop', {
    value: scrollTop,
    writable: true,
  });
  Object.defineProperty(el, 'scrollHeight', {
    value: scrollHeight,
    writable: true,
  });
  Object.defineProperty(el, 'clientHeight', {
    value: clientHeight,
    writable: true,
  });
  el.scrollTo = (options?: ScrollToOptions | number, y?: number) => {
    if (typeof options === 'number') {
      el.scrollTop = typeof y === 'number' ? y : options;
      return;
    }
    if (options?.top !== undefined) {
      el.scrollTop = options.top;
    }
  };
  return el;
}

const baseMessages = [createMessage(1, '2024-01-01T00:00:00Z')];

afterEach(() => {
  vi.restoreAllMocks();
});

describe('useScrollBehavior', () => {
  it('triggers fetchOlderHistory when near top with guards', () => {
    const fetchOlderHistory = vi.fn();
    const fetchOlderDmHistory = vi.fn();

    const { result } = renderHook(() =>
      useScrollBehavior({
        messages: baseMessages,
        isDm: false,
        roomId: 1,
        dmId: null,
        hasMore: true,
        isLoadingOlder: false,
        skipInitialScroll: true,
        fetchOlderHistory,
        fetchOlderDmHistory,
      })
    );

    const el = setupDom(10);
    result.current.listRef.current = el;

    act(() => {
      result.current.handleScroll();
    });

    expect(fetchOlderHistory).toHaveBeenCalledWith(1);
    expect(fetchOlderDmHistory).not.toHaveBeenCalled();
  });

  it('does not trigger load when already loading or no more', () => {
    const fetchOlderHistory = vi.fn();
    const fetchOlderDmHistory = vi.fn();

    const { result, rerender } = renderHook(
      ({ hasMore, isLoadingOlder }) =>
        useScrollBehavior({
          messages: baseMessages,
          isDm: false,
          roomId: 1,
          dmId: null,
          hasMore,
          isLoadingOlder,
          skipInitialScroll: true,
          fetchOlderHistory,
          fetchOlderDmHistory,
        }),
      { initialProps: { hasMore: false, isLoadingOlder: false } }
    );

    const el = setupDom(10);
    result.current.listRef.current = el;

    act(() => {
      result.current.handleScroll();
    });

    expect(fetchOlderHistory).not.toHaveBeenCalled();

    rerender({ hasMore: true, isLoadingOlder: true });

    act(() => {
      result.current.handleScroll();
    });

    expect(fetchOlderHistory).not.toHaveBeenCalled();
  });
});
