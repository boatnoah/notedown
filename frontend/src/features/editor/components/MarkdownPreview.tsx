import DOMPurify from "dompurify";
import { useEffect, useRef } from "react";
import { marked } from "marked";

type MarkdownPreviewProps = {
  markdown: string;
};

function renderHtml(el: HTMLElement, html: string) {
  el.innerHTML = DOMPurify.sanitize(html);
}

export function MarkdownPreview({ markdown }: MarkdownPreviewProps) {
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const el = ref.current;
    if (!el) {
      return;
    }

    const rendered = marked.parse(markdown);
    if (rendered instanceof Promise) {
      rendered.then((html) => renderHtml(el, html));
    } else {
      renderHtml(el, rendered);
    }
  }, [markdown]);

  return <div ref={ref} id="preview" />;
}
