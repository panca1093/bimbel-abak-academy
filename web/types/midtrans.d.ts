interface SnapSuccessResult {
  status_code?: string;
  transaction_status?: string;
  order_id?: string;
}

interface SnapCallbacks {
  onSuccess?: (result: SnapSuccessResult) => void;
  onPending?: (result: SnapSuccessResult) => void;
  onError?: (result: SnapSuccessResult) => void;
  onClose?: () => void;
}

interface Snap {
  pay: (
    token: string,
    callbacks?: SnapCallbacks,
  ) => void;
}

interface Window {
  snap?: Snap;
}