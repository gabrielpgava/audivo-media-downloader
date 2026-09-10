export namespace models {
	
	export class UserError {
	    code: string;
	    message: string;
	    retryable: boolean;
	    details?: string;
	
	    static createFrom(source: any = {}) {
	        return new UserError(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.message = source["message"];
	        this.retryable = source["retryable"];
	        this.details = source["details"];
	    }
	}
	export class ActionResponse {
	    ok: boolean;
	    value?: string;
	    error?: UserError;
	
	    static createFrom(source: any = {}) {
	        return new ActionResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ok = source["ok"];
	        this.value = source["value"];
	        this.error = this.convertValues(source["error"], UserError);
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
	export class HistoryItem {
	    id: string;
	    sourceUrl: string;
	    service: string;
	    title: string;
	    type: string;
	    filePath?: string;
	    outputDir: string;
	    completedAt: string;
	    status: string;
	    itemCount?: number;
	    completedItems?: number;
	    missing?: boolean;
	
	    static createFrom(source: any = {}) {
	        return new HistoryItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.sourceUrl = source["sourceUrl"];
	        this.service = source["service"];
	        this.title = source["title"];
	        this.type = source["type"];
	        this.filePath = source["filePath"];
	        this.outputDir = source["outputDir"];
	        this.completedAt = source["completedAt"];
	        this.status = source["status"];
	        this.itemCount = source["itemCount"];
	        this.completedItems = source["completedItems"];
	        this.missing = source["missing"];
	    }
	}
	export class MediaOption {
	    mediaType: string;
	    format: string;
	    quality?: string;
	    label: string;
	
	    static createFrom(source: any = {}) {
	        return new MediaOption(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.mediaType = source["mediaType"];
	        this.format = source["format"];
	        this.quality = source["quality"];
	        this.label = source["label"];
	    }
	}
	export class CollectionItem {
	    id?: string;
	    url: string;
	    title: string;
	    artist?: string;
	    album?: string;
	    index: number;
	    durationSeconds?: number;
	
	    static createFrom(source: any = {}) {
	        return new CollectionItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.url = source["url"];
	        this.title = source["title"];
	        this.artist = source["artist"];
	        this.album = source["album"];
	        this.index = source["index"];
	        this.durationSeconds = source["durationSeconds"];
	    }
	}
	export class MediaPreview {
	    url: string;
	    service: string;
	    kind: string;
	    title: string;
	    itemCount?: number;
	    items?: CollectionItem[];
	    thumbnailUrl?: string;
	    author?: string;
	    durationSeconds?: number;
	    mediaType: string;
	    options: MediaOption[];
	
	    static createFrom(source: any = {}) {
	        return new MediaPreview(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.url = source["url"];
	        this.service = source["service"];
	        this.kind = source["kind"];
	        this.title = source["title"];
	        this.itemCount = source["itemCount"];
	        this.items = this.convertValues(source["items"], CollectionItem);
	        this.thumbnailUrl = source["thumbnailUrl"];
	        this.author = source["author"];
	        this.durationSeconds = source["durationSeconds"];
	        this.mediaType = source["mediaType"];
	        this.options = this.convertValues(source["options"], MediaOption);
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
	export class AnalyzeResponse {
	    preview?: MediaPreview;
	    duplicate?: HistoryItem;
	    error?: UserError;
	
	    static createFrom(source: any = {}) {
	        return new AnalyzeResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.preview = this.convertValues(source["preview"], MediaPreview);
	        this.duplicate = this.convertValues(source["duplicate"], HistoryItem);
	        this.error = this.convertValues(source["error"], UserError);
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
	
	export class DiagnosticResponse {
	    report: string;
	    error?: UserError;
	
	    static createFrom(source: any = {}) {
	        return new DiagnosticResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.report = source["report"];
	        this.error = this.convertValues(source["error"], UserError);
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
	export class DownloadJob {
	    id: string;
	    state: string;
	    filename?: string;
	    message?: string;
	
	    static createFrom(source: any = {}) {
	        return new DownloadJob(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.state = source["state"];
	        this.filename = source["filename"];
	        this.message = source["message"];
	    }
	}
	export class DownloadRequest {
	    url: string;
	    title?: string;
	    mediaType: string;
	    format: string;
	    quality?: string;
	    outputDir: string;
	    collectionTitle?: string;
	    collectionItems?: CollectionItem[];
	
	    static createFrom(source: any = {}) {
	        return new DownloadRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.url = source["url"];
	        this.title = source["title"];
	        this.mediaType = source["mediaType"];
	        this.format = source["format"];
	        this.quality = source["quality"];
	        this.outputDir = source["outputDir"];
	        this.collectionTitle = source["collectionTitle"];
	        this.collectionItems = this.convertValues(source["collectionItems"], CollectionItem);
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
	export class EngineStatus {
	    name: string;
	    version?: string;
	    requiredVersion?: string;
	    path?: string;
	    available: boolean;
	    valid: boolean;
	    message?: string;
	
	    static createFrom(source: any = {}) {
	        return new EngineStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.version = source["version"];
	        this.requiredVersion = source["requiredVersion"];
	        this.path = source["path"];
	        this.available = source["available"];
	        this.valid = source["valid"];
	        this.message = source["message"];
	    }
	}
	export class EngineResponse {
	    engines: EngineStatus[];
	    error?: UserError;
	
	    static createFrom(source: any = {}) {
	        return new EngineResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.engines = this.convertValues(source["engines"], EngineStatus);
	        this.error = this.convertValues(source["error"], UserError);
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
	
	
	export class HistoryResponse {
	    items: HistoryItem[];
	
	    static createFrom(source: any = {}) {
	        return new HistoryResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.items = this.convertValues(source["items"], HistoryItem);
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
	
	
	export class ProviderAuthStatus {
	    provider: string;
	    state: string;
	    connected: boolean;
	    message?: string;
	    sessionId?: string;
	    lastCapturedUnix?: number;
	    error?: UserError;
	
	    static createFrom(source: any = {}) {
	        return new ProviderAuthStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.provider = source["provider"];
	        this.state = source["state"];
	        this.connected = source["connected"];
	        this.message = source["message"];
	        this.sessionId = source["sessionId"];
	        this.lastCapturedUnix = source["lastCapturedUnix"];
	        this.error = this.convertValues(source["error"], UserError);
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
	export class SettingsView {
	    downloadDirectory: string;
	    theme: string;
	    audioFormat: string;
	    videoQuality?: string;
	    quickMode?: boolean;
	    appleMusicConnected: boolean;
	    youtubeConnected: boolean;
	    lastEngineCheckUnix?: number;
	    warning?: string;
	
	    static createFrom(source: any = {}) {
	        return new SettingsView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.downloadDirectory = source["downloadDirectory"];
	        this.theme = source["theme"];
	        this.audioFormat = source["audioFormat"];
	        this.videoQuality = source["videoQuality"];
	        this.quickMode = source["quickMode"];
	        this.appleMusicConnected = source["appleMusicConnected"];
	        this.youtubeConnected = source["youtubeConnected"];
	        this.lastEngineCheckUnix = source["lastEngineCheckUnix"];
	        this.warning = source["warning"];
	    }
	}
	export class StartDownloadResponse {
	    job?: DownloadJob;
	    error?: UserError;
	
	    static createFrom(source: any = {}) {
	        return new StartDownloadResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.job = this.convertValues(source["job"], DownloadJob);
	        this.error = this.convertValues(source["error"], UserError);
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
	export class UpdateResponse {
	    currentVersion: string;
	    latestVersion?: string;
	    releaseUrl?: string;
	    available: boolean;
	    error?: UserError;
	
	    static createFrom(source: any = {}) {
	        return new UpdateResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.currentVersion = source["currentVersion"];
	        this.latestVersion = source["latestVersion"];
	        this.releaseUrl = source["releaseUrl"];
	        this.available = source["available"];
	        this.error = this.convertValues(source["error"], UserError);
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

