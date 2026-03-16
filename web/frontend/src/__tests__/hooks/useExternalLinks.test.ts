import { describe, expect, it } from 'vitest';
import { inferFileNameFromUrl } from '../../pages/Chat/components/MessageList/hooks/useExternalLinks';

describe('inferFileNameFromUrl', () => {
  it('decodes encoded filenames', () => {
    const url = 'https://example.com/files/%E6%96%87%E4%BB%B6.txt';
    expect(inferFileNameFromUrl(url)).toBe('??.txt');
  });

  it('falls back to raw segment on decode error', () => {
    const url = 'https://example.com/files/%E0%A4%A';
    expect(inferFileNameFromUrl(url)).toBe('%E0%A4%A');
  });

  it('handles empty path', () => {
    const url = 'https://example.com/';
    expect(inferFileNameFromUrl(url)).toBe('????');
  });
});
