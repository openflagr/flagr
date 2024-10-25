import axios from "axios"
import constants from '@/constants'
import store from '../store';  // Adjust the path to your store
import router from "../router";

let axiosInstance ;
let flagrAxiosInstance;

const API_URLS = constants.API_URLS
const API_URL = constants.API_URL
const SSO_URL = constants.SSO_URL
export const setupAxiosInstance = () => {
	axiosInstance = axios.create({
		baseURL: SSO_URL,
		timeout: 300000,
		headers: headerParams(),
	})
	//setupAxiosInstanceInterceptors()
	return axiosInstance
}

export const setupflagrAxiosInstance = () => {
	flagrAxiosInstance = axios.create({
		baseURL: API_URL,
		timeout: 300000
	})
	//setupAxiosInstanceInterceptors()
	return flagrAxiosInstance
}

const deviceType = ["web", "mweb", "ios-webview", "android-webview"]

export const detectDeviceType = () => {
	if (typeof window !== "undefined") {
		const details = navigator.userAgent
		const regexp = /android|iphone|kindle|ipad/i
		const isMobileDevice = regexp.test(details)
		if (isMobileDevice) {
			return deviceType[1]
		} else {
			return deviceType[0]
		}
	}
}

const headerParams = (default_headers = {}) => {
	const headers = default_headers
	headers["Content-Type"] = "application/json"
	headers["x-client-type"] = detectDeviceType() ?? "unknown"
	headers["accept"] = "application/json"
	if (typeof window !== "undefined") {
		let uuid = localStorage.getItem("uuid")
		if (!uuid) {
			uuid = generateUUID()
			localStorage?.setItem("uuid", uuid)
		}
		headers["x-device-id"] = uuid
	}
	return headers
}


export const generateUUID = () => {
	let d = new Date().getTime()
	let d2 = (typeof performance !== "undefined" && performance.now && performance.now() * 1000) || 0 //Time in microseconds since page-load or 0 if unsupported
	return "xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx".replace(/[xy]/g, function (c) {
		let r = Math.random() * 16 //random number between 0 and 16
		if (d > 0) {
			//Use timestamp until depleted
			r = (d + r) % 16 | 0
			d = Math.floor(d / 16)
		} else {
			//Use microseconds since page-load if supported
			r = (d2 + r) % 16 | 0
			d2 = Math.floor(d2 / 16)
		}
		return (c === "x" ? r : (r & 0x3) | 0x8).toString(16)
	})
}

export const getAxiosInstance = () => {
	if (!axiosInstance) {
		throw new Error("Axios instance not initialized")
	}
	const token = localStorage.getItem('tokens');

	// Try to parse the token and catch any potential errors
	if(token){
		try {
			const authToken = token ? JSON.parse(token) : null;
			if (authToken && authToken["x-access-token"]) {
				axiosInstance.defaults.headers["Authorization"] = "Bearer " + authToken["x-access-token"];
			}
		} catch (error) {
			console.error("Failed to parse token from localStorage", error);
		}
	} else {
		delete axiosInstance.defaults.headers["Authorization"]
	}

	// Set the Authorization header if the token exist

	return axiosInstance;
}


export const getAxiosFlagrInstance = () => {
	if (!flagrAxiosInstance) {
		throw new Error("Axios instance not initialized")
	}
	const token = localStorage.getItem('tokens');

	// Try to parse the token and catch any potential errors
	if(token){
		try {
			const authToken = token ? JSON.parse(token) : null;
			if (authToken && authToken["x-access-token"]) {
				flagrAxiosInstance.defaults.headers["Authorization"] = "Bearer " + authToken["x-access-token"];
			}
		} catch (error) {
			console.error("Failed to parse token from localStorage", error);
		}
	} else {
		delete flagrAxiosInstance.defaults.headers["Authorization"]
	}

	// Set the Authorization header if the token exist

	return flagrAxiosInstance;
}


export const getGoogleRedirectionLink = async (postData) => {
    const response = await getAxiosInstance().get(
		`${`api/v1/auth/authenticate/googleAuthRedirectLink`}?uiRedirectUrl=${postData.uiRedirectUrl}`,
	)
	const redirectUrlResponse = response?.data?.data
	return redirectUrlResponse
}


export const googleAuthenticate = async (postData) => {	
	const reqPayload = JSON.stringify(postData)
	try {
		const response = await getAxiosInstance().post(`api/v1/auth/authenticate/google`, reqPayload)
		if (typeof window !== "undefined") {
			const token = {}
			token["x-access-token"] = response.headers["x-access-token"]
			token["x-refresh-token"] = response.headers["x-refresh-token"]
			localStorage.setItem("tokens", JSON.stringify(token))
		}
		return response?.data?.data
	} catch (error) {
		console.error(error)
		if (typeof window !== "undefined") {
			localStorage.removeItem("tokens")
		}
		throw error
	}
}

export function isAuthenticated() {
	return !!localStorage.getItem('tokens'); // Assuming token is saved in localStorage
}

export async function logout() {
	
	try {
		await getAxiosInstance().post(`api/v1/auth/logout`)
	} catch (error) {
		alert('logging you out');
	} finally {
		store.dispatch("reset")
		if (typeof window !== "undefined") {
			localStorage.removeItem("tokens")
			router.replace({ name: 'Login' })
		}
	}
}

export async function getUserDetails() {
	const response = await getAxiosInstance().get(`${API_URLS.USER_DETAILS}`)
	return response.data.data
}