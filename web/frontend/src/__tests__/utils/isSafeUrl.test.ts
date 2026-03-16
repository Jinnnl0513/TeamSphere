import { describe, expect, it } from 'vitest';
import { isSafeUrl } from '../../pages/Chat/components/MessageList/MessageItem';

describe('isSafeUrl', () => {
  it('accepts http/https urls', () => {
    expect(isSafeUrl('http://example.com/a.png')).toBe(true);
    expect(isSafeUrl('https://example.com/a.png')).toBe(true);
  });

  it('rejects non-http protocols', () => {
    expect(isSafeUrl('javascript:alert(1)')).toBe(false);
    expect(isSafeUrl('data:text/plain,hello')).toBe(false);
    expect(isSafeUrl('ftp://example.com/file')).toBe(false);
  });

  it('handles relative urls safely', () => {
    expect(isSafeUrl('/uploads/a.png')).toBe(true);
  });

  it('handles empty input', () => {
    expect(isSafeUrl('')).toBe(false);
  });
});
