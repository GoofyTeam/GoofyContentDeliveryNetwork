import Input from "@/components/Input";
import { Link } from "@tanstack/react-router";
import React, { ChangeEvent, useState } from "react";

const Register: React.FC = () => {
  const [showPassword, setShowPassword] = useState(false);
  const [formData, setFormData] = useState({
    firstName: '',
    lastName: '',
    email: '',
    password: '',
    confirmPassword: '',
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
            Créer un compte
          </h2>
          <p className='mt-2 text-center text-sm text-gray-600'>
            Ou{' '}
            <Link
              to='/login'
              className='font-medium text-blue-600 hover:text-blue-500'
            >
              connectez-vous à votre compte existant
            </Link>
          </p>
        </div>

        <form className='mt-8 space-y-6' onSubmit={handleSubmit}>
          <div className='space-y-4'>
            <div className='grid grid-cols-1 gap-4 sm:grid-cols-2'>
              <Input
                label='Prénom'
                id='firstName'
                name='firstName'
                value={formData.firstName}
                onChange={handleChange}
                placeholder='John'
                required
              />

              <Input
                label='Nom'
                id='lastName'
                name='lastName'
                value={formData.lastName}
                onChange={handleChange}
                placeholder='Doe'
                required
              />
            </div>

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
              value={formData.password}
              onChange={handleChange}
              required
              showPasswordToggle
              placeholder='********'
              showPassword={showPassword}
              onTogglePassword={() => setShowPassword(!showPassword)}
            />

            <Input
              label='Confirmer le mot de passe'
              id='confirmPassword'
              name='confirmPassword'
              placeholder='********'
              type='password'
              value={formData.confirmPassword}
              onChange={handleChange}
              required
            />
          </div>

          <div>
            <button
              type='submit'
              className='group relative w-full flex justify-center py-2 px-4 border border-transparent text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500'
            >
              S'inscrire
            </button>
          </div>
        </form>
      </div>
    </section>
  );
};

export default Register;
