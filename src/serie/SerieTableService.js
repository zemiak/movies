import { RestClientService } from "../RestClientService.js";

export class SerieTableService extends RestClientService {
    constructor() {
        super();
    }

    getCustomEventName() {
        return "SerieTableService";
    }

    getServicePath() {
        return "/series/all";
    }

    getCacheKeyName() {
        return "_" + this.getCustomEventName();
    }

    setId(id) {
        this.id = id;
    }
}
