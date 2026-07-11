// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

export {
  E2EKeyPair,
  E2EChannel,
  deriveKeyFromSecret,
  deriveConfirmationCode,
  generateNonce,
  generateSecret,
  generateSessionSalt,
  SESSION_SALT_LEN,
  SESSION_SALT_ACK,
  createPairingRecord,
  deriveChannelFromRecord,
  type PairingRecord,
} from "./crypto.js";
export {
  connect,
  acceptSessionSalt,
  wakeRelay,
  Session,
  Stream,
  Datagram,
  encodeUvarint,
  decodeUvarint,
  encodeStreamHeader,
  decodeStreamHeader,
  encodeDatagram,
  decodeDatagram,
  type ConnectArgs,
  type Identity,
} from "./relay.js";
