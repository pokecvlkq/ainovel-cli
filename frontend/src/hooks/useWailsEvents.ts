import { useEffect } from 'react';
import { useNovelStore } from '../stores/novelStore';
// @ts-ignore
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime';

export function useWailsEvents() {
  const { setSnapshot, addEvent, appendStream, clearStream, setComplete } = useNovelStore();

  useEffect(() => {
    try {
      EventsOn('novel:snapshot', (data: any) => {
        setSnapshot(data);
      });

      EventsOn('novel:event', (data: any) => {
        addEvent(data);
      });

      EventsOn('novel:stream', (data: string) => {
        appendStream(data);
      });

      EventsOn('novel:stream-clear', () => {
        clearStream();
      });

      // Khi AI viết xong toàn bộ → cập nhật trạng thái hoàn thành
      EventsOn('novel:done', () => {
        setComplete(true);
      });

      return () => {
        EventsOff('novel:snapshot');
        EventsOff('novel:event');
        EventsOff('novel:stream');
        EventsOff('novel:stream-clear');
        EventsOff('novel:done');
      };
    } catch (e) {
      console.warn("Wails runtime not found. Events disabled.");
    }
  }, [setSnapshot, addEvent, appendStream, clearStream, setComplete]);
}
