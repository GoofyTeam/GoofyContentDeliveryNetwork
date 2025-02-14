import { X } from 'lucide-react';
import axios from 'axios';
import { useState } from 'react';

interface CreateFolderModalProps {
  isOpen: boolean;
  onClose: () => void;
  onCreateFolder: (folderName: string) => void;
}

const CreateFolderModal = ({ isOpen, onClose, onCreateFolder }: CreateFolderModalProps) => {
  const [folderName, setFolderName] = useState('');
  const [error, setError] = useState<string | null>(null);

  const api = axios.create({
    baseURL: 'http://localhost:8082/api',
    headers: {
      Authorization: 'Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6InRlc3Q0QGdtYWlsLmNvbSIsImV4cCI6MTczOTYyODAwNCwidXNlcl9pZCI6IjY3YWY0NzVmZjI0ZWVlZjFmYWI2ZmU5NSJ9.LrLYhffochKEgF8PJhK-S7uH6_WhFkmrkPtWEja1pl0',
      'Content-Type': 'application/json',
    },
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!folderName.trim()) {
      setError('Le nom du dossier est requis');
      return;
    }

    try {
      const response = await api.post('/folders', {
        name: folderName
      });

      console.log('Dossier créé:', response.data);
      
      onCreateFolder(folderName);
      setFolderName('');
      setError(null);
      onClose();
    } catch (err) {
      if (axios.isAxiosError(err)) {
        setError(err.response?.data?.message || 'Erreur lors de la création du dossier');
      } else {
        setError('Une erreur inattendue est survenue');
      }
    }
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 overflow-y-auto">
      <div className="flex min-h-screen items-center justify-center p-4">
        <div
          className="fixed inset-0 bg-black/30 backdrop-blur-sm"
          onClick={onClose}
        />
        
        <div className="relative bg-white dark:bg-gray-800 rounded-lg shadow-xl max-w-md w-full mx-4">
          <div className="flex items-center justify-between p-4 border-b dark:border-gray-700">
            <h3 className="text-lg font-semibold text-gray-900 dark:text-white">
              Nouveau dossier
            </h3>
            <button
              onClick={onClose}
              className="text-gray-400 hover:text-gray-500 dark:hover:text-gray-300"
            >
              <X className="h-5 w-5" />
            </button>
          </div>

          <form onSubmit={handleSubmit} className="p-6">
            <div className="space-y-4">
              <div>
                <label
                  htmlFor="folder-name"
                  className="block text-sm font-medium text-gray-700 dark:text-gray-300"
                >
                  Nom du dossier
                </label>
                <input
                  type="text"
                  id="folder-name"
                  value={folderName}
                  onChange={(e) => setFolderName(e.target.value)}
                  className="mt-1 block w-full rounded-md border border-gray-300 dark:border-gray-600 
                           shadow-sm py-2 px-3 bg-white dark:bg-gray-700 
                           text-gray-900 dark:text-white
                           focus:border-blue-500 focus:ring-1 focus:ring-blue-500"
                  placeholder="Mon nouveau dossier"
                />
              </div>

              {error && (
                <p className="text-sm text-red-600 dark:text-red-400">
                  {error}
                </p>
              )}
            </div>

            <div className="mt-6 flex justify-end gap-3">
              <button
                type="button"
                onClick={onClose}
                className="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-200 
                         hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg"
              >
                Annuler
              </button>
              <button
                type="submit"
                className="px-4 py-2 text-sm font-medium text-white bg-blue-600 
                         hover:bg-blue-700 rounded-lg"
              >
                Créer
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  );
};

export default CreateFolderModal;