import { useEffect, useRef } from "react";
import { marked } from "marked";

type MarkdownPreviewProps = {
  markdown: string;
};

export function MarkdownPreview({ markdown }: MarkdownPreviewProps) {
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const el = ref.current;
    if (!el) {
      return;
    }

    const rendered = marked.parse(markdown);
    if (rendered instanceof Promise) {
      rendered.then((html) => {
        el.innerHTML = html;
      });
    } else {
      el.innerHTML = rendered;
    }
  }, [markdown]);

  return <div ref={ref} id="preview" />;
}
