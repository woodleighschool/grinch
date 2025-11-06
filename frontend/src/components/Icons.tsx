import {
  BarChart3,
  Settings,
  Users,
  Smartphone,
  Shield,
  User,
  TreePine,
  Laptop,
  DoorClosedLocked,
  AppWindowMac,
  ChevronUp,
  ChevronDown,
  ChevronsUpDown,
  ChevronLeft,
  ChevronRight,
  BookOpen,
  CandyCane,
} from "lucide-react";

// Icon wrapper for consistent styling
const Icon = ({ children, className = "" }: { children: React.ReactNode; className?: string }) => (
  <span className={`icon ${className}`}>{children}</span>
);

export const Icons = {
  Brand: () => (
    <Icon className="brand-icon">
      <TreePine size={20} />
    </Icon>
  ),
  Dashboard: () => (
    <Icon className="nav-icon">
      <BarChart3 size={16} />
    </Icon>
  ),
  Applications: () => (
    <Icon className="nav-icon">
      <AppWindowMac size={16} />
    </Icon>
  ),
  Users: () => (
    <Icon className="nav-icon">
      <Users size={16} />
    </Icon>
  ),
  Settings: () => (
    <Icon className="settings-icon">
      <Settings size={16} />
    </Icon>
  ),
  Shield: ({ className = "" }: { className?: string } = {}) => (
    <Icon className={className}>
      <Shield size={16} />
    </Icon>
  ),
  Devices: () => (
    <Icon className="nav-icon">
      <Laptop size={16} />
    </Icon>
  ),
  User: () => (
    <Icon>
      <User size={16} />
    </Icon>
  ),
  Group: () => (
    <Icon>
      <Users size={16} />
    </Icon>
  ),

  Logout: () => (
    <Icon>
      <DoorClosedLocked size={16} />
    </Icon>
  ),
  Help: () => (
    <Icon>
      <BookOpen size={16} />
    </Icon>
  ),
  CandyCane: () => (
    <Icon>
      <CandyCane size={16} />
    </Icon>
  ),
  ChevronUp: ({ size = 16 }: { size?: number } = {}) => <ChevronUp size={size} />,
  ChevronDown: ({ size = 16 }: { size?: number } = {}) => <ChevronDown size={size} />,
  ChevronsUpDown: ({ size = 16 }: { size?: number } = {}) => <ChevronsUpDown size={size} />,
  ChevronLeft: ({ size = 16 }: { size?: number } = {}) => <ChevronLeft size={size} />,
  ChevronRight: ({ size = 16 }: { size?: number } = {}) => <ChevronRight size={size} />,
};
