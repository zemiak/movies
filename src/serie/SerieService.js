import { RestClientService } from "../RestClientService.js";

export class SerieService extends RestClientService {
    constructor() {
        super();
    }

    getCustomEventName() {
        return "SerieService";
    }

    getServicePath() {
        return "/series/browse?id=" + this.id;
    }

    getCacheKeyName() {
        return "_" + this.getCustomEventName() + "_" + this.id;
    }

    setId(id) {
        this.id = id;
    }
}
