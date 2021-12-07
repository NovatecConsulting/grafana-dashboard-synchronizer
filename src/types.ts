import { DataSourceJsonData, DataSourceSettings } from '@grafana/data';

/**
 * These are options configured for each DataSource instance.
 */
export interface SynchronizeOptions extends DataSourceJsonData {
  grafanaUrl?: string;
  gitUrl?: string;

  pushConfiguration?: PushOptions;
  pullConfiguration?: PullOptions;
}

export interface PushPullOptions {
  enable?: boolean;
  gitBranch?: string;
  syncInterval?: number;
  filter?: string;
}

export interface PushOptions extends PushPullOptions {
  tagPattern?: string;
  pushTags?: boolean;
}

export interface PullOptions extends PushPullOptions {}

/**
 * Value that is used in the backend, but never sent over HTTP to the frontend
 */
export interface SecureSynchronizeOptions {
  grafanaApiToken?: string;
  privateSshKey?: string;
}

export interface OptionsChange {
  options: DataSourceSettings<SynchronizeOptions, SecureSynchronizeOptions>;
  onChange: (newOptions: DataSourceSettings<SynchronizeOptions, SecureSynchronizeOptions>) => void;
}

export interface PushPullOptionsChange {
  ppOptions: PushPullOptions;
  onChange: (newOptions: PushPullOptions) => void;
}
