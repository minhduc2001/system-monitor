export interface MenuItem {
  key: string;
  label: string;
  icon?: React.ReactNode;
  path?: string;
  children?: MenuItem[];
  permission?: string;
  roles?: string[];
  hidden?: boolean;
}

export interface Permission {
  id: string;
  name: string;
  code: string;
  description?: string;
}

export interface Role {
  id: string;
  name: string;
  code: string;
  permissions: string[];
  description?: string;
}
