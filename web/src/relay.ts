// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

/** Options for relay connections. */
export interface ConnectOptions {
  /** Bearer token for authentication on /register. */
  token?: string;
  /**
   * WebSocket constructor, for Node.js compatibility.
   * Defaults to globalThis.WebSocket.
   */
  WebSocket?: new (url: string, protocols?: string | string[]) => WebSocket;
}

/** Queued incoming message. */
interface QueuedMessage {
  data: Uint8Array | string;
  resolve: (value: Uint8Array | string) => void;
}

/**
 * A connection to a peer through the tern relay.
 */
export class Conn {
  /** The relay-assigned instance ID. */
  public instanceID: string;
  /** Called when the WebSocket closes. */
  public onclose: ((event: CloseEvent) => void) | null = null;

  private ws: WebSocket;
  private recvQueue: (Uint8Array | string)[] = [];
  private waiters: ((value: Uint8Array | string) => void)[] = [];
  private closed = false;
  private closeError: Error | null = null;

  /** @internal Use register() or connect() instead. */
  constructor(ws: WebSocket, instanceID: string) {
    this.ws = ws;
    this.instanceID = instanceID;
    ws.binaryType = "arraybuffer";

    ws.onmessage = (event: MessageEvent) => {
      const data =
        event.data instanceof ArrayBuffer
          ? new Uint8Array(event.data)
          : (event.data as string);

      if (this.waiters.length > 0) {
        const waiter = this.waiters.shift()!;
        waiter(data);
      } else {
        this.recvQueue.push(data);
      }
    };

    ws.onclose = (event: CloseEvent) => {
      this.closed = true;
      this.closeError = new Error(
        `WebSocket closed: code=${event.code} reason=${event.reason}`,
      );
      // Reject all pending waiters
      for (const waiter of this.waiters) {
        // We use a special rejection mechanism below
      }
      this.waiters = [];
      if (this.onclose) {
        this.onclose(event);
      }
    };
  }

  /** Send a message to the peer. */
  send(data: Uint8Array | string): void {
    if (this.closed) {
      throw new Error("connection is closed");
    }
    this.ws.send(data);
  }

  /** Receive the next message from the peer. */
  recv(): Promise<Uint8Array | string> {
    if (this.recvQueue.length > 0) {
      return Promise.resolve(this.recvQueue.shift()!);
    }
    if (this.closed) {
      return Promise.reject(this.closeError ?? new Error("connection closed"));
    }
    return new Promise<Uint8Array | string>((resolve, reject) => {
      // Wrap to handle close-while-waiting
      const originalOnClose = this.ws.onclose;
      this.waiters.push(resolve);

      // If the socket closes while we're waiting, reject
      const idx = this.waiters.length - 1;
      const checkClose = this.ws.addEventListener("close", () => {
        const waiterIdx = this.waiters.indexOf(resolve);
        if (waiterIdx >= 0) {
          this.waiters.splice(waiterIdx, 1);
          reject(
            this.closeError ?? new Error("connection closed while waiting"),
          );
        }
      }, { once: true });
    });
  }

  /** Close the connection. */
  close(): void {
    if (!this.closed) {
      this.ws.close(1000, "");
    }
  }
}

/**
 * Convert an HTTP(S) URL to a WebSocket URL.
 */
function toWsURL(url: string): string {
  if (url.startsWith("http://")) return "ws://" + url.slice(7);
  if (url.startsWith("https://")) return "wss://" + url.slice(8);
  if (url.startsWith("ws://") || url.startsWith("wss://")) return url;
  return url;
}

/**
 * Register as a backend with the relay. Returns a Conn whose instanceID
 * is the relay-assigned instance ID.
 */
export function register(
  relayURL: string,
  opts?: ConnectOptions,
): Promise<Conn> {
  return new Promise((resolve, reject) => {
    const WS = opts?.WebSocket ?? globalThis.WebSocket;
    const wsURL = toWsURL(relayURL) + "/register";

    // WebSocket API doesn't support arbitrary headers in browsers.
    // Use the subprotocol field to pass the token if needed.
    const protocols: string[] = [];
    if (opts?.token) {
      protocols.push("Bearer-" + opts.token);
    }

    const ws = protocols.length > 0 ? new WS(wsURL, protocols) : new WS(wsURL);
    ws.binaryType = "arraybuffer";

    let gotID = false;

    ws.onmessage = (event: MessageEvent) => {
      if (!gotID) {
        gotID = true;
        const id =
          typeof event.data === "string"
            ? event.data
            : new TextDecoder().decode(event.data);
        const conn = new Conn(ws, id);
        resolve(conn);
      }
    };

    ws.onerror = () => {
      if (!gotID) {
        reject(new Error(`WebSocket connection to ${wsURL} failed`));
      }
    };

    ws.onclose = (event: CloseEvent) => {
      if (!gotID) {
        reject(
          new Error(
            `WebSocket closed before instance ID received: code=${event.code}`,
          ),
        );
      }
    };
  });
}

/**
 * Connect to a backend instance through the relay.
 */
export function connect(
  relayURL: string,
  instanceID: string,
  opts?: ConnectOptions,
): Promise<Conn> {
  return new Promise((resolve, reject) => {
    const WS = opts?.WebSocket ?? globalThis.WebSocket;
    const wsURL = toWsURL(relayURL) + "/ws/" + instanceID;

    const ws = new WS(wsURL);
    ws.binaryType = "arraybuffer";

    ws.onopen = () => {
      resolve(new Conn(ws, instanceID));
    };

    ws.onerror = () => {
      reject(new Error(`WebSocket connection to ${wsURL} failed`));
    };

    ws.onclose = (event: CloseEvent) => {
      reject(
        new Error(`WebSocket closed before open: code=${event.code}`),
      );
    };
  });
}
