import Input from "@/components/Input";
import { Link } from "@tanstack/react-router";
import React, { ChangeEvent, useState } from "react";

const Login: React.FC = () => {
  const [showPassword, setShowPassword] = useState(false);
  const [formData, setFormData] = useState({
    email: '',
    password: '',
  });

  const handleChange = (e: ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setFormData((prev) => ({
      ...prev,
      [name]: value,
    }));
  };

  const handleSubmit = (e: ChangeEvent<HTMLFormElement>) => {
    e.preventDefault();
    console.log('Form submitted:', formData);
  };

  return (
    <section className='min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8'>
      <div className='max-w-md w-full space-y-8 bg-white p-8 rounded-lg shadow-md'>
        <div>
          <h2 className='mt-6 text-center text-3xl font-extrabold text-gray-900'>
            Connexion
          </h2>
          <p className='mt-2 text-center text-sm text-gray-600'>
            Ou{' '}
            <Link
              to='/register'
              className='font-medium text-blue-600 hover:text-blue-500'
            >
              créez un nouveau compte
            </Link>
          </p>
        </div>

        <form className='mt-8 space-y-6' onSubmit={handleSubmit}>
          <div className='space-y-4'>
            <Input
              label='Adresse email'
              id='email'
              name='email'
              type='email'
              value={formData.email}
              onChange={handleChange}
              required
              autoComplete='email'
              placeholder='exemple@email.com'
            />

            <Input
              label='Mot de passe'
              id='password'
              name='password'
              placeholder='********'
              value={formData.password}
              onChange={handleChange}
              required
              showPasswordToggle
              showPassword={showPassword}
              onTogglePassword={() => setShowPassword(!showPassword)}
              autoComplete='current-password'
            />

            <div className='flex items-center justify-between'>
              <div className='flex items-center'>
                <input
                  id='remember-me'
                  name='remember-me'
                  type='checkbox'
                  className='h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded'
                />
                <label
                  htmlFor='remember-me'
                  className='ml-2 block text-sm text-gray-900'
                >
                  Se souvenir de moi
                </label>
              </div>

              <div className='text-sm'>
                <a
                  href='#'
                  className='font-medium text-blue-600 hover:text-blue-500'
                >
                  Mot de passe oublié ?
                </a>
              </div>
            </div>
          </div>

          <div>
            <button
              type='submit'
              className='group relative w-full flex justify-center py-2 px-4 border border-transparent text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500'
            >
              Se connecter
            </button>
          </div>
        </form>
      </div>
    </section>
  );
};

export default Login;
