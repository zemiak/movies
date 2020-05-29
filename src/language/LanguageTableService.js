import { RestClientService } from "../RestClientService.js";

export class LanguageTableService extends RestClientService {
    constructor() {
        super();
    }

    getCustomEventName() {
        return "LanguageTableService";
    }

    getServicePath() {
        return "/languages/all";
    }

    getCacheKeyName() {
        return "_" + this.getCustomEventName();
    }

    setId(id) {
        this.id = id;
    }
}
