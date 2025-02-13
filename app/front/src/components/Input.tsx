import React from 'react';
import { Eye, EyeOff } from 'lucide-react';

interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  label?: string;
  error?: string;
  showPasswordToggle?: boolean;
  showPassword?: boolean;
  onTogglePassword?: () => void;
}

const Input = ({
  label,
  error,
  showPasswordToggle = false,
  showPassword,
  onTogglePassword,
  className = '',
  type = 'text',
  id,
  ...props
}: InputProps) => {
  const inputProps = {
    id,
    type: showPasswordToggle ? (showPassword ? 'text' : 'password') : type,
    className: `
      block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm
      focus:outline-none focus:ring-blue-500 focus:border-blue-500
      ${error ? 'border-red-500' : ''}
      ${className}
    `,
    ...props,
  };

  return (
    <div className='w-full'>
      {label && (
        <label
          htmlFor={id}
          className='block text-sm font-medium text-gray-700 mb-1'
        >
          {label}
        </label>
      )}

      <div className='relative'>
        <input {...inputProps} />

        {showPasswordToggle && (
          <button
            type='button'
            className='absolute inset-y-0 right-0 pr-3 flex items-center'
            onClick={onTogglePassword}
          >
            {showPassword ? (
              <EyeOff className='h-5 w-5 text-gray-400' />
            ) : (
              <Eye className='h-5 w-5 text-gray-400' />
            )}
          </button>
        )}
      </div>

      {error && <p className='mt-1 text-sm text-red-600'>{error}</p>}
    </div>
  );
};

export default Input;
