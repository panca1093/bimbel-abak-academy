// Polls for window.snap availability, waiting up to timeoutMs milliseconds.
// Resolves true once available, false on timeout.
export function waitForSnap(timeoutMs = 10000): Promise<boolean> {
  const start = Date.now();
  return new Promise((resolve) => {
    function check() {
      if (typeof window !== "undefined" && window.snap) {
        resolve(true);
        return;
      }
      if (Date.now() - start >= timeoutMs) {
        resolve(false);
        return;
      }
      setTimeout(check, 100);
    }
    check();
  });
}
