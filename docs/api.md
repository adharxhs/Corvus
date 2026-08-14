# API Reference

Base URL: `http://localhost:8080` (or `https://<host>` when exposed via tunnel).

## Authentication

All routes below require an `Authorization: Bearer <JWT>` header unless noted otherwise.

---

## Public routes

### `GET /`

Health check. Returns `{"status": "ok"}`.

### `GET /health`

Same as `GET /`.

### `POST /register`

Create a new account.

**Request body**:
```json
{ "username": "alice", "password": "s3cret" }
```

**Response (201)**:
```json
{ "token": "eyJ..." }
```

**Errors**: `400` malformed body, `409` username taken.

### `POST /login`

Authenticate and receive a JWT.

**Request body**:
```json
{ "username": "alice", "password": "s3cret" }
```

**Response (200)**:
```json
{ "token": "eyJ..." }
```

**Errors**: `400` malformed body, `401` invalid credentials.

---

## User routes

### `GET /users/by-username/{username}`

Exact-username lookup. Returns only the user ID (no other data).

**Response (200)**:
```json
{ "id": "uuid" }
```

**Errors**: `404` user not found.

### `GET /users/{id}`

Resolve a user ID to its username.

**Response (200)**:
```json
{ "id": "uuid", "username": "alice" }
```

**Errors**: `404` user not found.

---

## Relationship routes (chat requests)

### `POST /chat-request`

Send a chat request to another user. Fails if a pending request already exists in either direction, or if the cooldown has not elapsed since a previous rejection.

**Request body**:
```json
{ "recipient_id": "uuid" }
```

**Response (201)**:
```json
{ "status": "pending" }
```

**Errors**: `400` bad body, `409` duplicate/cooldown active, `404` user not found.

### `GET /chat-requests`

List incoming pending chat requests for the authenticated user.

**Response (200)**:
```json
[
  { "requester_id": "uuid", "status": "pending", "created_at": 1234567890 }
]
```

### `POST /chat-request/{requester_id}/accept`

Accept a pending chat request. This is bidirectional — both users become "accepted."

**Response (200)**: `{"status": "accepted"}`

### `POST /chat-request/{requester_id}/reject`

Reject a pending chat request. Silent — no notification to the requester.

**Response (200)**: `{"status": "rejected"}`

---

## Group routes

### `POST /groups`

Create a new group. The creator is automatically added as a member.

**Request body**:
```json
{ "name": "Project X" }
```

**Response (201)**:
```json
{ "group_id": "uuid", "name": "Project X" }
```

### `GET /groups/invites`

List pending group invites for the authenticated user.

**Response (200)**:
```json
[
  { "invite_id": "uuid", "group_id": "uuid", "group_name": "Project X", "inviter_id": "uuid" }
]
```

### `GET /groups/{group_id}/members`

List members of a group (must be a member).

**Response (200)**:
```json
[
  { "user_id": "uuid", "username": "alice" }
]
```

### `POST /groups/{group_id}/invite`

Invite a user to a group. Requires an accepted personal-chat relationship between inviter and invitee, and the inviter must be a group member.

**Request body**:
```json
{ "user_id": "uuid" }
```

### `POST /groups/{group_id}/invite/accept`

Accept a group invite. The invitee becomes a group member.

### `DELETE /groups/{group_id}/member`

Leave a group unilaterally.

---

## Prekey routes

### `POST /prekey`

Upsert your own prekey bundle (identity key, signed prekey, optional one-time prekeys).

**Request body**:
```json
{
  "identity_key": "base64...",
  "signed_prekey": "base64...",
  "signed_prekey_signature": "base64...",
  "one_time_prekeys": [
    { "id": 1, "public_key": "base64..." }
  ]
}
```

### `GET /prekey/{id}`

Fetch another user's prekey bundle by user ID. **Ungated** — does not require an accepted relationship. X3DH session establishment is decoupled from the accept step.

**Response (200)**: The prekey bundle object.

---

## Profile picture routes

### `POST /profile-picture`

Upload an encrypted profile picture. Must be a newer version than the currently stored one.

**Request body**:
```json
{
  "ciphertext": "base64...",
  "nonce": "base64...",
  "version": 2
}
```

**Response (204)**: No body.

### `GET /profile-picture/{id}`

Get an encrypted profile picture by user ID. Requires an accepted relationship with that user.

**Response (200)**:
```json
{
  "ciphertext": "base64...",
  "nonce": "base64...",
  "version": 2
}
```

---

## WebSocket

### `GET /ws?token=<JWT>`

Upgrade to WebSocket. Token is passed as a query parameter (HTTP Upgrade cannot carry custom headers in all clients).

Server-originated control messages:

| Type | Payload | Description |
|---|---|---|
| `presence_snapshot` | `{"online": ["user_id", ...]}` | Sent immediately after connect; lists which accepted contacts are currently online |
| `presence` | `{"user_id": "...", "status": "online\|offline"}` | Broadcast to online accepted contacts on every connect/disconnect |

Client→server message types (via the same WebSocket):

| Type | Payload |
|---|---|
| `direct_message` | `{"recipient_id": "...", "ciphertext": "...", "ratchet_header": "..."}` |
| `group_message` | `{"group_id": "...", "ciphertext": "...", "key_id": ..., "counter": ...}` |
| `sender_key_distribution` | `{"group_id": "...", "distribution": "..."}` |
| `profile_picture_updated` | `{"version": 2}` |
