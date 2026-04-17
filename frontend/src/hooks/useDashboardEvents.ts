import { useEffect, useState } from 'react';
import {
  startEventService,
  disconnect,
  onStatusChange,
  type ConnectionStatus,
} from '../services/eventService';

/**
 * Manages the SSE event service lifecycle.
 *
 * - Connects to the SSE stream on mount
 * - Disconnects on unmount
 * - Returns the current connection status
 *
 * Usage:
 *   const status = useDashboardEvents();
 */
export function useDashboardEvents(): ConnectionStatus {
  const [status, setStatus] = useState<ConnectionStatus>('disconnected');

  useEffect(() => {
    startEventService();

    const unsubscribe = onStatusChange(setStatus);

    return () => {
      unsubscribe();
      disconnect();
    };
  }, []);

  return status;
}
