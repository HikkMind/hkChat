import { useEffect, useRef } from 'react';

export default function useScrollToBottom(dependencies = []) {
  const ref = useRef(null);

  useEffect(() => {
    if (ref.current) {
      ref.current.scrollTop = ref.current.scrollHeight;
    }
  }, dependencies);

  return ref;
}