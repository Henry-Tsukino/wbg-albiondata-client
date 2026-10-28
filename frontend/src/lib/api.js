// Thin wrappers around the generated Wails bindings.
//
// The generated bindings call Go methods by a numeric ID. If the frontend
// bundle and the Go binary were built from different binding snapshots the
// IDs don't match and Wails answers "unknown bound method id N". Calling by
// the fully qualified method name has no such dependency, so it is used
// first; the generated ID-based call is kept only as a fallback.
import { Call } from '@wailsio/runtime';
import { DashboardService } from '../../bindings/github.com/Henry-Tsukino/wbg-albiondata-client/internal/dashboard/index.js';

const SERVICE = 'github.com/Henry-Tsukino/wbg-albiondata-client/internal/dashboard.DashboardService';

export async function setMarketDisabled(disabled) {
  try {
    await Call.ByName(`${SERVICE}.SetMarketDisabled`, disabled);
  } catch (err) {
    console.warn('SetMarketDisabled by name failed, trying by ID:', err);
    await DashboardService.SetMarketDisabled(disabled);
  }
}
