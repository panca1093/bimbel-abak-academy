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

interface SnapEmbedOptions extends SnapCallbacks {
  embedId: string;
}

interface Snap {
  pay: (
    token: string,
    callbacks?: SnapCallbacks,
  ) => void;
  embed: (
    token: string,
    options: SnapEmbedOptions,
  ) => void;
}

interface Window {
  snap?: Snap;
}
