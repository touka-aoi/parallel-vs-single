// ゲームループ・状態管理

import type { Actor } from "./protocol";
import {
  CONTROL_SUBTYPE_ASSIGN,
  CONTROL_SUBTYPE_JOIN,
  CONTROL_SUBTYPE_LEAVE,
  DATA_TYPE_ACTOR,
  DATA_TYPE_CONTROL,
  decodeActorBroadcast,
  decodeAssignMessage,
  encodeControlMessage,
  encodeInputMessage,
  getControlSubType,
  getDataType,
  sessionIdToString,
} from "./protocol";
import { WebSocketClient } from "./websocket";
import { InputManager } from "./input";
import { Renderer } from "./renderer";

const SERVER_URL = "ws://localhost:9090/ws";

export class Game {
  private ws: WebSocketClient;
  private input: InputManager;
  private renderer: Renderer;

  private actors: Actor[] = [];
  private mySessionId: Uint8Array | null = null;
  private seq: number = 0;
  private connected: boolean = false;

  constructor(canvas: HTMLCanvasElement) {
    this.input = new InputManager();
    this.renderer = new Renderer(canvas);
    this.ws = new WebSocketClient(
      SERVER_URL,
      this.onMessage.bind(this),
      this.onConnect.bind(this),
      this.onDisconnect.bind(this)
    );
  }

  start(): void {
    this.ws.connect();
    this.gameLoop();
  }

  private onConnect(): void {
    this.connected = true;
    console.log("Connected to server, waiting for session ID...");
    // Assignメッセージを待つ（ここではJoinを送信しない）
  }

  private onDisconnect(): void {
    this.connected = false;
    this.actors = [];
    this.mySessionId = null;
    console.log("Disconnected from server");
  }

  private onMessage(data: ArrayBuffer): void {
    const dataType = getDataType(data);

    if (dataType === DATA_TYPE_CONTROL) {
      const subType = getControlSubType(data);
      if (subType === CONTROL_SUBTYPE_ASSIGN) {
        // セッションID通知を受信
        this.mySessionId = decodeAssignMessage(data);
        console.log("Received session ID:", sessionIdToString(this.mySessionId));

        // Joinメッセージを送信
        const joinMsg = encodeControlMessage(this.mySessionId, this.seq++, CONTROL_SUBTYPE_JOIN);
        this.ws.send(joinMsg);
        console.log("Sent Join message");
      }
    } else if (dataType === DATA_TYPE_ACTOR) {
      this.actors = decodeActorBroadcast(data);
    }
  }

  private gameLoop(): void {
    // 入力送信
    if (this.connected && this.mySessionId !== null) {
      const keyMask = this.input.getKeyMask();
      if (keyMask !== 0) {
        const msg = encodeInputMessage(this.mySessionId, this.seq++, keyMask);
        this.ws.send(msg);
      }
    }

    // 描画
    this.renderer.render(this.actors, this.mySessionId);

    requestAnimationFrame(this.gameLoop.bind(this));
  }

  destroy(): void {
    // Control/Leave を送信（ベストエフォート）
    if (this.connected && this.mySessionId !== null) {
      const leaveMsg = encodeControlMessage(this.mySessionId, this.seq++, CONTROL_SUBTYPE_LEAVE);
      this.ws.send(leaveMsg);
    }
    this.ws.disconnect();
    this.input.destroy();
  }
}