export interface Incident {
  id: string;
  title: string;
  description: string;
  status: string;
  started_at: string;
  resolved_at?: string;
}
