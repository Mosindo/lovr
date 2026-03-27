import { useSyncExternalStore } from "react";

type FeedbackState = {
  error: string | null;
  loadingCount: number;
  loadingLabel: string | null;
};

type Listener = () => void;

let state: FeedbackState = {
  error: null,
  loadingCount: 0,
  loadingLabel: null
};

const listeners = new Set<Listener>();

function emit() {
  listeners.forEach((listener) => listener());
}

function subscribe(listener: Listener) {
  listeners.add(listener);
  return () => {
    listeners.delete(listener);
  };
}

function getSnapshot(): FeedbackState {
  return state;
}

export function beginGlobalLoading(label = "Working...") {
  state = {
    ...state,
    loadingCount: state.loadingCount + 1,
    loadingLabel: label
  };
  emit();
}

export function endGlobalLoading() {
  const nextCount = Math.max(0, state.loadingCount - 1);
  state = {
    ...state,
    loadingCount: nextCount,
    loadingLabel: nextCount === 0 ? null : state.loadingLabel
  };
  emit();
}

export function showGlobalError(message: string) {
  if (!message) {
    return;
  }
  state = {
    ...state,
    error: message
  };
  emit();
}

export function clearGlobalError() {
  if (state.error === null) {
    return;
  }
  state = {
    ...state,
    error: null
  };
  emit();
}

export function useGlobalFeedback() {
  return useSyncExternalStore(subscribe, getSnapshot, getSnapshot);
}
