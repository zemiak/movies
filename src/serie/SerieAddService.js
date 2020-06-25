import { RestClientService } from "../RestClientService.js";

export class SerieAddService extends RestClientService {
    getCustomEventName() {
        return "SerieAddService";
    }

    getServicePath() {
        return "/series/detail/forNew";
    }
}
