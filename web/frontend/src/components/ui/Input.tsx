import React from 'react';
import { cn } from '../../utils/cn';

interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {
    error?: boolean;
}

export const Input = React.forwardRef<HTMLInputElement, InputProps>(
    ({ className, type, error, ...props }, ref) => {
        return (
            <input
                type={type}
                className={cn(
                    'flex h-10 w-full rounded-md bg-[var(--bg-input)] px-3 py-2 text-sm text-[var(--text-main)] transition-colors placeholder:text-[var(--text-muted)] focus-visible:outline-none focus:ring-2 focus:ring-[var(--accent)] disabled:cursor-not-allowed disabled:opacity-50 shadow-inner',
                    error && 'border-[var(--color-discord-red-500)] focus-visible:ring-[var(--color-discord-red-500)]',
                    className
                )}
                ref={ref}
                {...props}
            />
        );
    }
);
Input.displayName = 'Input';
