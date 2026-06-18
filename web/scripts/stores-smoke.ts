import assert from "node:assert/strict";
import { useAuthStore } from "../stores/auth.ts";
import { useCartStore } from "../stores/cart.ts";

// Auth store shape (Task 6)
const auth = useAuthStore.getState();
assert.equal(typeof auth.clear, "function", "auth.clear must be a function");
assert.equal(typeof auth.setSession, "function", "auth.setSession must be a function");
assert.equal(auth.token, null, "token starts null");
assert.equal(auth.user, null, "user starts null");

// setSession mutates state
useAuthStore.getState().setSession("t1", { id: "u1", name: "A" });
const afterSet = useAuthStore.getState();
assert.equal(afterSet.token, "t1", "setSession stores token");
assert.deepEqual(afterSet.user, { id: "u1", name: "A" }, "setSession stores user");

// clear resets both
useAuthStore.getState().clear();
const afterClear = useAuthStore.getState();
assert.equal(afterClear.token, null, "clear resets token");
assert.equal(afterClear.user, null, "clear resets user");

// Cart store shape (UI-only badge mirror)
const cart = useCartStore.getState();
assert.equal(typeof cart.setCount, "function", "cart.setCount must be a function");
assert.equal(typeof cart.count, "number", "cart.count must be a number");
useCartStore.getState().setCount(3);
assert.equal(useCartStore.getState().count, 3, "setCount updates count");

console.log("smoke: stores OK");