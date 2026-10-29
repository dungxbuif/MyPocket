import * as React from 'react';

// Native select boundary for existing form-row styling and keyboard semantics.
export const Select = React.forwardRef<HTMLSelectElement, React.SelectHTMLAttributes<HTMLSelectElement>>(
  (props, ref) => <select ref={ref} {...props} />,
);
Select.displayName = 'Select';
