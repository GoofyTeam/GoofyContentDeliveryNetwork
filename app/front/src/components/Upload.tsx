import { FileData } from '@/pages/Index';
import { Loader, Upload, X } from 'lucide-react';
import { useState } from 'react';

const UploadModal = ({
  isOpen,
  onClose,
  onUploadComplete,
}: {
  isOpen: boolean;
  onClose: () => void;
  onUploadComplete: (newFiles: FileData[]) => void;
}) => {
  const [isDragging, setIsDragging] = useState(false);
  const [files, setFiles] = useState<File[]>([]);
  const [uploading, setUploading] = useState(false);

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(true);
  };

  const handleDragLeave = (e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(false);
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(false);
    const droppedFiles = Array.from(e.dataTransfer.files);
    setFiles((prev) => [...prev, ...droppedFiles]);
  };

  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files) {
      const selectedFiles = Array.from(e.target.files);
      setFiles((prev) => [...prev, ...selectedFiles]);
    }
  };

  const removeFile = (index: number) => {
    setFiles((prev) => prev.filter((_, i) => i !== index));
  };

  const handleUpload = async () => {
    setUploading(true);

    // Simulate upload and create new file entries
    await new Promise((resolve) => setTimeout(resolve, 2000));

    const newFiles: FileData[] = files.map((file, index) => ({
      id: Date.now() + index,
      name: file.name,
      type: file.name.split('.').pop() || 'unknown',
      size: `${(file.size / (1024 * 1024)).toFixed(1)} MB`,
      lastModified: new Date().toISOString().split('T')[0],
    }));

    onUploadComplete(newFiles);
    setUploading(false);
    setFiles([]);
    onClose();
  };

  if (!isOpen) return null;

  return (
    <div className='fixed inset-0 z-50 overflow-y-auto'>
      <div className='flex min-h-screen items-center justify-center p-4'>
        {/* Backdrop */}
        <div
          className='fixed inset-0 bg-black/30 backdrop-blur-sm'
          onClick={onClose}
        />

        {/* Modal */}
        <div className='relative bg-white dark:bg-gray-800 rounded-lg shadow-xl max-w-2xl w-full mx-4'>
          <div className='flex items-center justify-between p-4 border-b dark:border-gray-700'>
            <h3 className='text-lg font-semibold text-gray-900 dark:text-white'>
              Importer des fichiers
            </h3>
            <button
              onClick={onClose}
              className='text-gray-400 hover:text-gray-500 dark:hover:text-gray-300'
            >
              <X className='h-5 w-5' />
            </button>
          </div>

          <div className='p-6'>
            <div
              onDragOver={handleDragOver}
              onDragLeave={handleDragLeave}
              onDrop={handleDrop}
              className={`
                border-2 border-dashed rounded-lg p-8 text-center
                ${
                  isDragging
                    ? 'border-blue-500 bg-blue-50 dark:bg-blue-900/20'
                    : 'border-gray-300 dark:border-gray-600'
                }
              `}
            >
              <Upload className='mx-auto h-12 w-12 text-gray-400' />
              <p className='mt-4 text-sm text-gray-600 dark:text-gray-300'>
                Glissez-déposez vos fichiers ici, ou{' '}
                <label className='text-blue-500 hover:text-blue-600 cursor-pointer'>
                  parcourez
                  <input
                    type='file'
                    multiple
                    className='hidden'
                    onChange={handleFileSelect}
                  />
                </label>
              </p>
            </div>

            {files.length > 0 && (
              <div className='mt-6 space-y-2'>
                <h4 className='text-sm font-medium text-gray-900 dark:text-white'>
                  Fichiers sélectionnés
                </h4>
                <div className='max-h-48 overflow-y-auto'>
                  {files.map((file, index) => (
                    <div
                      key={index}
                      className='flex items-center justify-between py-2 px-3 bg-gray-50 dark:bg-gray-700/50 rounded'
                    >
                      <span className='text-sm text-gray-600 dark:text-gray-300 truncate'>
                        {file.name}
                      </span>
                      <button
                        onClick={() => removeFile(index)}
                        className='text-gray-400 hover:text-gray-500 dark:hover:text-gray-300'
                      >
                        <X className='h-4 w-4' />
                      </button>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>

          <div className='flex items-center justify-end gap-3 p-4 border-t dark:border-gray-700'>
            <button
              onClick={onClose}
              className='px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-200 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg'
            >
              Annuler
            </button>
            <button
              onClick={handleUpload}
              disabled={files.length === 0 || uploading}
              className={`
                px-4 py-2 text-sm font-medium text-white rounded-lg
                ${
                  files.length === 0 || uploading
                    ? 'bg-blue-400 cursor-not-allowed'
                    : 'bg-blue-600 hover:bg-blue-700'
                }
                flex items-center gap-2
              `}
            >
              {uploading ? (
                <>
                  <Loader className='h-4 w-4 animate-spin' />
                  Importation...
                </>
              ) : (
                'Importer'
              )}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

export default UploadModal;
