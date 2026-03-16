import React from 'react';
import { cn } from '../../utils/cn';

type LabelProps = React.LabelHTMLAttributes<HTMLLabelElement>;

export const Label = React.forwardRef<HTMLLabelElement, LabelProps>(
    ({ className, ...props }, ref) => (
        <label
            ref={ref}
            className={cn(
                'text-[12px] font-bold uppercase tracking-wide text-[var(--text-muted)] mb-2 block',
                className
            )}
            {...props}
        />
    )
);
Label.displayName = 'Label';
