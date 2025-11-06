import * as Dialog from "@radix-ui/react-dialog";
import { X } from "lucide-react";
import { ReactNode } from "react";
import { Button } from "./Button";

interface ConfirmDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title: string;
  description: string;
  onConfirm: () => void;
  confirmText?: string;
  cancelText?: string;
  destructive?: boolean;
}

export function ConfirmDialog({
  open,
  onOpenChange,
  title,
  description,
  onConfirm,
  confirmText = "Confirm",
  cancelText = "Cancel",
  destructive = false,
}: ConfirmDialogProps) {
  return (
    <Dialog.Root open={open} onOpenChange={onOpenChange}>
      <Dialog.Portal>
        <Dialog.Overlay className="dialog-overlay" />
        <Dialog.Content className="dialog-content">
          <Dialog.Title className="dialog-title">{title}</Dialog.Title>
          <Dialog.Description className="dialog-description">{description}</Dialog.Description>

          <div className="dialog-actions">
            <Dialog.Close asChild>
              <Button variant="secondary">{cancelText}</Button>
            </Dialog.Close>
            <Button
              variant={destructive ? "danger" : "primary"}
              onClick={() => {
                onConfirm();
                onOpenChange(false);
              }}
            >
              {confirmText}
            </Button>
          </div>

          <Dialog.Close asChild>
            <Button variant="ghost" className="dialog-close" aria-label="Close">
              <X size={16} />
            </Button>
          </Dialog.Close>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}

interface DialogContentProps {
  children: ReactNode;
  title?: string;
  description?: string;
  className?: string;
}

export function DialogRoot({ children, ...props }: { children: ReactNode } & Dialog.DialogProps) {
  return <Dialog.Root {...props}>{children}</Dialog.Root>;
}

export function DialogTrigger({ children, ...props }: { children: ReactNode } & Dialog.DialogTriggerProps) {
  return <Dialog.Trigger {...props}>{children}</Dialog.Trigger>;
}

export function DialogContent({ children, title, description, className = "" }: DialogContentProps) {
  return (
    <Dialog.Portal>
      <Dialog.Overlay className="dialog-overlay" />
      <Dialog.Content className={`dialog-content ${className}`}>
        {title && <Dialog.Title className="dialog-title">{title}</Dialog.Title>}
        {description && <Dialog.Description className="dialog-description">{description}</Dialog.Description>}
        {children}
        <Dialog.Close asChild>
          <Button variant="ghost" className="dialog-close" aria-label="Close">
            <X size={16} />
          </Button>
        </Dialog.Close>
      </Dialog.Content>
    </Dialog.Portal>
  );
}

export function DialogClose({ children, ...props }: { children: ReactNode } & Dialog.DialogCloseProps) {
  return <Dialog.Close {...props}>{children}</Dialog.Close>;
}
