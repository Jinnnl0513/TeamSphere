import { describe, expect, it, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { renderTextWithMentionsAndLinks } from '../../pages/Chat/components/MessageList/MessageItem';

describe('renderTextWithMentionsAndLinks', () => {
  it('renders mentions and links', async () => {
    const onOpenLinkConfirm = vi.fn();
    const content = 'hi @alice https://example.com/page';

    render(
      <div>
        {renderTextWithMentionsAndLinks(content, 'alice', onOpenLinkConfirm)}
      </div>
    );

    expect(screen.getByText('@alice')).toBeInTheDocument();
    const linkButton = screen.getByRole('button', { name: 'https://example.com/page' });
    expect(linkButton).toBeInTheDocument();

    await userEvent.click(linkButton);
    expect(onOpenLinkConfirm).toHaveBeenCalledWith('https://example.com/page');
  });
});
