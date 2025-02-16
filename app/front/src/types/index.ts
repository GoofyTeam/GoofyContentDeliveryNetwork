export interface Folder {
  id: string;
  name: string;
  path: string;
  parent_id?: string;
  user_id: string;
  depth: number;
  created_at: string;
  updated_at: string;
}