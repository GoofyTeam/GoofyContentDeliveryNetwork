import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import useAuth from "@/hooks/useAuth";
import { useNavigate, useParams, useRouter } from "@tanstack/react-router";
import { useEffect, useState } from "react";
import { Button } from "./ui/button";
import { PlusIcon } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "./ui/dialog";
import { Input } from "./ui/input";

interface FolderItem {
  id: string;
  name: string;
}

export function SiteHeader() {
  const navigate = useNavigate();
  const router = useRouter();
  const params = useParams({
    from: "/_auth/drive/$folderPath",
  });

  const folderPath = params?.folderPath ?? undefined;
  const { logout, isAuth, user, accessToken } = useAuth();
  const [email, setEmail] = useState<string | null>(null);
  const [folderName, setFolderName] = useState("");
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [isFileDialogOpen, setIsFileDialogOpen] = useState(false);

  useEffect(() => {
    if (!isAuth) {
      navigate({ to: "/login" });
    } else {
      setEmail(user?.email ?? "");
    }
  }, [isAuth, navigate, user?.email]);

  const handleLogout = async () => {
    await logout();

    navigate({ to: "/login" });
  };

  const handleCreateFolder = async () => {
    if (!folderName.trim()) return;

    let parentId = folderPath;
    if (!parentId) {
      try {
        const foldersResponse = await fetch(
          "http://localhost:8082/api/folders",
          {
            method: "GET",
            headers: {
              Authorization: `Bearer ${accessToken}`,
              "Content-Type": "application/json",
            },
          }
        );
        if (!foldersResponse.ok) {
          throw new Error("Error recuperating folder");
        }
        const folders = await foldersResponse.json();
        const rootFolder = folders.find(
          (folder: { id: string; name: string }) =>
            folder.name.toLowerCase() === "root"
        );
        if (!rootFolder) {
          throw new Error("Root folder not found");
        }
        parentId = rootFolder.name;
      } catch (error) {
        console.error(error);
        return;
      }
    }

    try {
      const response = await fetch("http://localhost:8082/api/folders", {
        method: "POST",
        headers: {
          Authorization: `Bearer ${accessToken}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          name: folderName,
          parent_id: parentId,
        }),
      });

      if (!response.ok) {
        throw new Error("Error creating folder");
      }

      router.invalidate();

      setFolderName("");
      setIsDialogOpen(false);
    } catch (error) {
      console.error(error);
    }
  };

  const handleNewFile = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();

    const target = e.target as typeof e.target & {
      file: { files: FileList };
    };

    const file = target.file.files?.[0];
    if (!file) {
      throw Error('No file found')
    }

    let currentfolderName = folderPath;
    if (!currentfolderName) {
      try {
        const foldersResponse = await fetch(
          "http://localhost:8082/api/folders",
          {
            method: "GET",
            headers: {
              Authorization: `Bearer ${accessToken}`,
              "Content-Type": "application/json",
            },
          }
        );

        if (!foldersResponse.ok) {
          throw new Error("Error recuperating file");
        }

        const folders: FolderItem[] = await foldersResponse.json();
        const rootFolder = folders.find(
          (folder) => folder.name.toLowerCase() === "root"
        );

        if (!rootFolder) {
          throw new Error("Root folder not found");
        }

        currentfolderName = rootFolder.name;
      } catch (error) {
        console.error(error);
        return;
      }
    }

    const formData = new FormData();
    formData.append("file", file);
    formData.append("folder_id", currentfolderName);


    const res = await fetch("http://localhost:8082/api/files", {
      method: "POST",
      headers: {
        Authorization: `Bearer ${accessToken}`,
      },
      body: formData,
    });

    if (res.ok) {
      setIsFileDialogOpen(false);
      router.invalidate();
    } else {
      throw new Error('Error when sending file')
    }
  };

  return (
    <header className="border-grid top-0 z-50 w-full bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60 p-3 flex items-center justify-between">
      <div className="flex items-center gap-4">
        <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
          <DialogTrigger asChild>
            <Button variant="outline">
              <PlusIcon /> New Folder
            </Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Create a new folder</DialogTitle>
              <DialogDescription className="mt-6">
                <Input
                  placeholder="Folder name"
                  value={folderName}
                  onChange={(e) => setFolderName(e.target.value)}
                />
              </DialogDescription>
            </DialogHeader>
            <div className="flex justify-end gap-2">
              <Button variant="ghost" onClick={() => setIsDialogOpen(false)}>
                Cancel
              </Button>
              <Button
                onClick={handleCreateFolder}
                disabled={!folderName.trim()}
              >
                Create Folder
              </Button>
            </div>
          </DialogContent>
        </Dialog>
        <Dialog open={isFileDialogOpen} onOpenChange={setIsFileDialogOpen}>
          <DialogTrigger asChild>
            <Button variant="outline">
              <PlusIcon /> New File
            </Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Add a new file</DialogTitle>
              <DialogDescription className="my-2">
                Please select a file to upload
              </DialogDescription>
            </DialogHeader>
            <form onSubmit={handleNewFile} className="flex gap-4">
              <Input
                name="file"
                type="file"
                placeholder="File name"
                className="cursor-pointer"
              />
              <Button type="submit">Save file</Button>
            </form>
          </DialogContent>
        </Dialog>
      </div>
      <Popover>
        <PopoverTrigger>
          <Avatar>
            <AvatarFallback>
              {email ? email.slice(0, 2).toUpperCase() : "U"}
            </AvatarFallback>
          </Avatar>
        </PopoverTrigger>
        <PopoverContent>
          <div className="flex flex-col items-center gap-2">
            {email ? email : "Utilisateur"}
            <button
              onClick={handleLogout}
              className="mt-4 px-4 py-2 bg-red-500 text-white rounded hover:bg-red-600"
            >
              Déconnexion
            </button>
          </div>
        </PopoverContent>
      </Popover>
    </header>
  );
}
