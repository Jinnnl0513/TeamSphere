import React from 'react';
import { cn } from '../../utils/cn';

interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
    variant?: 'primary' | 'secondary' | 'danger' | 'ghost';
    size?: 'sm' | 'md' | 'lg';
}

export function Button({
    className,
    variant = 'primary',
    size = 'md',
    ...props
}: ButtonProps) {
    return (
        <button
            className={cn(
                'inline-flex items-center justify-center rounded-md font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 disabled:pointer-events-none disabled:opacity-50',
                {
                    'bg-[var(--accent)] text-white hover:bg-[var(--accent-hover)] active:bg-[#3b47bf] focus-visible:ring-blue-400': variant === 'primary',
                    'bg-[var(--bg-input)] text-[var(--text-main)] hover:bg-[#4e5058] active:bg-[#5b5d66] focus-visible:ring-gray-400': variant === 'secondary',
                    'bg-[var(--color-discord-red-500)] text-white hover:bg-[var(--color-discord-red-600)] active:bg-[#c22d31] focus-visible:ring-red-400': variant === 'danger',
                    'bg-transparent hover:bg-[var(--bg-input)] text-[var(--text-muted)] hover:text-[var(--text-main)]': variant === 'ghost',

                    'h-8 px-3 text-sm': size === 'sm',
                    'h-10 px-4 py-2 text-sm': size === 'md',
                    'h-12 px-8 text-base': size === 'lg',
                },
                className
            )}
            {...props}
        />
    );
}
