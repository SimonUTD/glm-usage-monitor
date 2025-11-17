export namespace main {
	
	export class Database {
	    // Go type: sql
	    DB?: any;
	
	    static createFrom(source: any = {}) {
	        return new Database(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.DB = this.convertValues(source["DB"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace models {
	
	export class APIToken {
	    id: number;
	    token_name: string;
	    token_value: string;
	    is_active: boolean;
	    created_at: time.Time;
	    updated_at: time.Time;
	
	    static createFrom(source: any = {}) {
	        return new APIToken(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.token_name = source["token_name"];
	        this.token_value = source["token_value"];
	        this.is_active = source["is_active"];
	        this.created_at = this.convertValues(source["created_at"], time.Time);
	        this.updated_at = this.convertValues(source["updated_at"], time.Time);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class AutoSyncConfig {
	    id: number;
	    config_key: string;
	    config_value: string;
	    description?: string;
	    updated_at: time.Time;
	
	    static createFrom(source: any = {}) {
	        return new AutoSyncConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.config_key = source["config_key"];
	        this.config_value = source["config_value"];
	        this.description = source["description"];
	        this.updated_at = this.convertValues(source["updated_at"], time.Time);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class BillFilter {
	    page_num: number;
	    page_size: number;
	    start_date?: time.Time;
	    end_date?: time.Time;
	    model_name?: string;
	    charge_type?: string;
	    group_name?: string;
	    min_cash_cost?: number;
	    max_cash_cost?: number;
	    search_term?: string;
	
	    static createFrom(source: any = {}) {
	        return new BillFilter(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.page_num = source["page_num"];
	        this.page_size = source["page_size"];
	        this.start_date = this.convertValues(source["start_date"], time.Time);
	        this.end_date = this.convertValues(source["end_date"], time.Time);
	        this.model_name = source["model_name"];
	        this.charge_type = source["charge_type"];
	        this.group_name = source["group_name"];
	        this.min_cash_cost = source["min_cash_cost"];
	        this.max_cash_cost = source["max_cash_cost"];
	        this.search_term = source["search_term"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ChargeTypeStatsData {
	    charge_type: string;
	    call_count: number;
	    cash_cost: number;
	    percentage: number;
	
	    static createFrom(source: any = {}) {
	        return new ChargeTypeStatsData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.charge_type = source["charge_type"];
	        this.call_count = source["call_count"];
	        this.cash_cost = source["cash_cost"];
	        this.percentage = source["percentage"];
	    }
	}
	export class ExpenseBill {
	    id: number;
	    charge_name: string;
	    charge_type: string;
	    model_name: string;
	    use_group_name: string;
	    group_name: string;
	    discount_rate: number;
	    cost_rate: number;
	    cash_cost: number;
	    billing_no: string;
	    order_time: string;
	    use_group_id: string;
	    group_id: string;
	    charge_unit: number;
	    charge_count: number;
	    charge_unit_symbol: string;
	    trial_cash_cost: number;
	    transaction_time: time.Time;
	    time_window_start: time.Time;
	    time_window_end: time.Time;
	    time_window: string;
	    create_time: time.Time;
	
	    static createFrom(source: any = {}) {
	        return new ExpenseBill(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.charge_name = source["charge_name"];
	        this.charge_type = source["charge_type"];
	        this.model_name = source["model_name"];
	        this.use_group_name = source["use_group_name"];
	        this.group_name = source["group_name"];
	        this.discount_rate = source["discount_rate"];
	        this.cost_rate = source["cost_rate"];
	        this.cash_cost = source["cash_cost"];
	        this.billing_no = source["billing_no"];
	        this.order_time = source["order_time"];
	        this.use_group_id = source["use_group_id"];
	        this.group_id = source["group_id"];
	        this.charge_unit = source["charge_unit"];
	        this.charge_count = source["charge_count"];
	        this.charge_unit_symbol = source["charge_unit_symbol"];
	        this.trial_cash_cost = source["trial_cash_cost"];
	        this.transaction_time = this.convertValues(source["transaction_time"], time.Time);
	        this.time_window_start = this.convertValues(source["time_window_start"], time.Time);
	        this.time_window_end = this.convertValues(source["time_window_end"], time.Time);
	        this.time_window = source["time_window"];
	        this.create_time = this.convertValues(source["create_time"], time.Time);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class HourlyUsageData {
	    hour: number;
	    call_count: number;
	    token_usage: number;
	    cash_cost: number;
	
	    static createFrom(source: any = {}) {
	        return new HourlyUsageData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.hour = source["hour"];
	        this.call_count = source["call_count"];
	        this.token_usage = source["token_usage"];
	        this.cash_cost = source["cash_cost"];
	    }
	}
	export class MembershipTierLimit {
	    id: number;
	    tier_name: string;
	    daily_limit?: number;
	    monthly_limit?: number;
	    max_tokens?: number;
	    max_context_length?: number;
	    features?: string;
	    description?: string;
	    updated_at: time.Time;
	
	    static createFrom(source: any = {}) {
	        return new MembershipTierLimit(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.tier_name = source["tier_name"];
	        this.daily_limit = source["daily_limit"];
	        this.monthly_limit = source["monthly_limit"];
	        this.max_tokens = source["max_tokens"];
	        this.max_context_length = source["max_context_length"];
	        this.features = source["features"];
	        this.description = source["description"];
	        this.updated_at = this.convertValues(source["updated_at"], time.Time);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ModelDistributionData {
	    model_name: string;
	    call_count: number;
	    token_usage: number;
	    cash_cost: number;
	    percentage: number;
	
	    static createFrom(source: any = {}) {
	        return new ModelDistributionData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.model_name = source["model_name"];
	        this.call_count = source["call_count"];
	        this.token_usage = source["token_usage"];
	        this.cash_cost = source["cash_cost"];
	        this.percentage = source["percentage"];
	    }
	}
	export class PaginationParams {
	    page: number;
	    size: number;
	    total: number;
	    has_next: boolean;
	
	    static createFrom(source: any = {}) {
	        return new PaginationParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.page = source["page"];
	        this.size = source["size"];
	        this.total = source["total"];
	        this.has_next = source["has_next"];
	    }
	}
	export class PaginatedResult {
	    data: any;
	    pagination: PaginationParams;
	
	    static createFrom(source: any = {}) {
	        return new PaginatedResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.data = source["data"];
	        this.pagination = this.convertValues(source["pagination"], PaginationParams);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class SyncStatus {
	    is_syncing: boolean;
	    last_sync_time?: time.Time;
	    last_sync_status?: string;
	    progress: number;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new SyncStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.is_syncing = source["is_syncing"];
	        this.last_sync_time = this.convertValues(source["last_sync_time"], time.Time);
	        this.last_sync_status = source["last_sync_status"];
	        this.progress = source["progress"];
	        this.message = source["message"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class StatsResponse {
	    total_records: number;
	    total_cash_cost: number;
	    hourly_usage: HourlyUsageData[];
	    model_distribution: ModelDistributionData[];
	    charge_type_stats: ChargeTypeStatsData[];
	    recent_usage: ExpenseBill[];
	    sync_status: SyncStatus;
	    membership_info?: MembershipTierLimit;
	
	    static createFrom(source: any = {}) {
	        return new StatsResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.total_records = source["total_records"];
	        this.total_cash_cost = source["total_cash_cost"];
	        this.hourly_usage = this.convertValues(source["hourly_usage"], HourlyUsageData);
	        this.model_distribution = this.convertValues(source["model_distribution"], ModelDistributionData);
	        this.charge_type_stats = this.convertValues(source["charge_type_stats"], ChargeTypeStatsData);
	        this.recent_usage = this.convertValues(source["recent_usage"], ExpenseBill);
	        this.sync_status = this.convertValues(source["sync_status"], SyncStatus);
	        this.membership_info = this.convertValues(source["membership_info"], MembershipTierLimit);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace services {
	
	export class APIService {
	
	
	    static createFrom(source: any = {}) {
	        return new APIService(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}
	export class SyncResult {
	    success: boolean;
	    message: string;
	    total_items: number;
	    synced_items: number;
	    failed_items: number;
	    skipped_items: number;
	    duration: number;
	    error_message?: string;
	    processed_bills?: models.ExpenseBill[];
	
	    static createFrom(source: any = {}) {
	        return new SyncResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	        this.total_items = source["total_items"];
	        this.synced_items = source["synced_items"];
	        this.failed_items = source["failed_items"];
	        this.skipped_items = source["skipped_items"];
	        this.duration = source["duration"];
	        this.error_message = source["error_message"];
	        this.processed_bills = this.convertValues(source["processed_bills"], models.ExpenseBill);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace time {
	
	export class Time {
	
	
	    static createFrom(source: any = {}) {
	        return new Time(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}

}

