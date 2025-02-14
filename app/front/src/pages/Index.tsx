import UploadModal from '@/components/Upload';
import { Download, Upload } from 'lucide-react';
import { useState } from 'react';

export interface FileData {
  id: number;
  name: string;
  type: string;
  size: string;
  lastModified: string;
  url?: string;
}

const initialFiles = [
  {
    id: 1,
    name: 'Document.pdf',
    type: 'pdf',
    size: '2.5 MB',
    lastModified: '2024-02-13',
    url: '/files/Document.pdf'
  },
  {
    id: 2,
    name: 'Images',
    type: 'folder',
    size: '128 MB',
    lastModified: '2024-02-12'
  },
  {
    id: 3,
    name: 'Projet.docx',
    type: 'docx',
    size: '1.8 MB',
    lastModified: '2024-02-11',
    url: '/files/Document.pdf'
  },
  {
    id: 4,
    name: 'Présentation.pptx',
    type: 'pptx',
    size: '5.2 MB',
    lastModified: '2024-02-10',
    url: '/files/Document.pdf'
  },
];

export default function Drive() {
  const [isUploadModalOpen, setIsUploadModalOpen] = useState(false);
  const [files, setFiles] = useState<FileData[]>(initialFiles);

  const handleUploadComplete = (newFiles: FileData[]) => {
    setFiles((prev) => [...newFiles, ...prev]);
  };

  const handleDownload = async (file: FileData) => {
    if (file.type === 'folder') {
      return;
    }

    try {
      if (!file.url) {
        throw new Error('URL de téléchargement non disponible');
      }

      const link = document.createElement('a');
      link.href = file.url;
      link.download = file.name;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
    } catch (error) {
      console.error('Erreur lors du téléchargement:', error);
    }
  };

  return (
    <div className='min-h-screen bg-gray-50 dark:bg-gray-900'>
      {/* Main content */}
      <main className='max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8'>
        {/* Actions bar */}
        <div className='flex items-center justify-between mb-8'>
          <button
            onClick={() => setIsUploadModalOpen(true)}
            className='flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700'
          >
            <Upload className='h-5 w-5' />
            Importer
          </button>
        </div>

        <div className='bg-white dark:bg-gray-800 rounded-lg shadow'>
          <table className='min-w-full divide-y divide-gray-200 dark:divide-gray-700'>
            <thead className='bg-gray-50 dark:bg-gray-900'>
              <tr>
                <th className='px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider'>
                  Nom
                </th>
                <th className='px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider'>
                  Type
                </th>
                <th className='px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider'>
                  Taille
                </th>
                <th className='px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider'>
                  Dernière modification
                </th>
                <th className='px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider'>
                  Actions
                </th>
              </tr>
            </thead>
            <tbody className='bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700'>
              {files.map((file) => (
                <tr
                  key={file.id}
                  className='hover:bg-gray-50 dark:hover:bg-gray-700 cursor-pointer'
                >
                  <td className='px-6 py-4 whitespace-nowrap'>
                    <div className='flex items-center'>
                      <img
                        src={`/${file.type}.svg`}
                        alt={file.type}
                        width={20}
                        height={20}
                        className='mr-3'
                      />
                      <span className='text-sm text-gray-900 dark:text-white'>
                        {file.name}
                      </span>
                    </div>
                  </td>
                  <td className='px-6 py-4 whitespace-nowrap'>
                    <span className='text-sm text-gray-500 dark:text-gray-400'>
                      {file.type}
                    </span>
                  </td>
                  <td className='px-6 py-4 whitespace-nowrap'>
                    <span className='text-sm text-gray-500 dark:text-gray-400'>
                      {file.size}
                    </span>
                  </td>
                  <td className='px-6 py-4 whitespace-nowrap'>
                    <span className='text-sm text-gray-500 dark:text-gray-400'>
                      {file.lastModified}
                    </span>
                  </td>
                  <td className='px-6 py-4 whitespace-nowrap'>
                    <button 
                      onClick={() => handleDownload(file)}
                      disabled={file.type === 'folder'}
                      className={`p-2 rounded-full hover:bg-gray-100 dark:hover:bg-gray-700 ${
                        file.type === 'folder' ? 'opacity-50 cursor-not-allowed' : 'cursor-pointer'
                      }`}
                      title={file.type === 'folder' ? 'Les dossiers ne peuvent pas être téléchargés' : 'Télécharger'}
                    >
                      <Download className='h-5 w-5 text-gray-500 dark:text-gray-400' />
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </main>

      <UploadModal
        isOpen={isUploadModalOpen}
        onClose={() => setIsUploadModalOpen(false)}
        onUploadComplete={handleUploadComplete}
      />
    </div>
  );
}