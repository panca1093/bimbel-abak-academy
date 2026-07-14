declare module "katex/contrib/auto-render" {
  interface AutoRenderDelimiters {
    left: string;
    right: string;
    display: boolean;
  }
  interface AutoRenderOptions {
    delimiters?: AutoRenderDelimiters[];
    ignoredTags?: string[];
    ignoredClasses?: string[];
    errorCallback?: (msg: string, err: unknown) => void;
    preProcess?: (math: string) => string;
    throwOnError?: boolean;
  }
  const renderMathInElement: (
    element: HTMLElement,
    options?: AutoRenderOptions,
  ) => void;
  export default renderMathInElement;
}
