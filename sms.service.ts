import { Injectable, Logger } from '@nestjs/common';
import axios, { AxiosInstance } from 'axios';
import envars from 'src/config/env-vars';
import { formatPhone } from 'src/shared/phone.util';

/**
 * SMS API configuration constants.
 */
const SMS_CONFIG = {
    LIVE_URL: 'https://www.egosms.co/api/v1/json/',
    SANDBOX_URL: 'http://sandbox.egosms.co/api/v1/json/',
    DEFAULT_PRIORITY: '1',
};

/**
 * SMS message payload type for SMS.
 */
export interface SmsMessage {
    number: string;
    message: string;
    senderid: string;
    priority?: string;
}

/**
 * SMS API response type.
 */
export interface SmsResponse {
    Status: string;
    Message: string;
    [key: string]: any;
}

/**
 * Service for sending SMS via SMS API.
 */
@Injectable()
export class SmsService {
    private readonly apiUrl: string;
    private readonly username: string;
    private readonly password: string;
    private readonly senderId: string;
    private readonly axios: AxiosInstance;
    private readonly logger = new Logger(SmsService.name);

    constructor() {
        this.username = envars.SMS_USERNAME;
        this.password = envars.SMS_PASSWORD;
        this.senderId = envars.SMS_SENDER_ID;
        this.apiUrl = SMS_CONFIG.LIVE_URL;
        this.axios = axios.create({
            timeout: 10000,
            headers: { 'Content-Type': 'application/json' },
        });
    }

    /**
     * Sends one or more SMS messages via SMS API.
     * @param messages Array of SmsMessage
     * @returns SmsResponse
     */
    async sendSms(messages: SmsMessage[]): Promise<SmsResponse> {
        const payload = {
            method: 'SendSms',
            userdata: {
                username: this.username,
                password: this.password,
            },
            msgdata: messages.map((msg) => ({
                number: formatPhone(msg.number),
                message: msg.message,
                senderid: msg.senderid || this.senderId,
                priority: msg.priority || SMS_CONFIG.DEFAULT_PRIORITY,
            })),
        };
        try {
            const response = await this.axios.post(this.apiUrl, payload);
            this.logger.log(`SMS response: ${JSON.stringify(response.data)}`);
            return response.data;
        } catch (error: any) {
            if (error.response) {
                this.logger.error(
                    `SMS HTTP error: ${error.response.status} - ${JSON.stringify(error.response.data)}`
                );
                return {
                    Status: 'Failed',
                    Message: `HTTP ${error.response.status}: ${JSON.stringify(error.response.data)}`,
                };
            } else if (error.request) {
                this.logger.error('SMS network error - no response received');
                return {
                    Status: 'Failed',
                    Message: 'Network error - check your connection',
                };
            } else {
                this.logger.error(`SMS error: ${error.message}`);
                return {
                    Status: 'Failed',
                    Message: error.message,
                };
            }
        }
    }

    /**
     * Sends a single SMS message.
     * @param number Recipient phone number
     * @param message Message text
     * @param senderid Sender ID (optional)
     * @param priority Priority (optional)
     */
    async sendSingleSms({ number, message, senderid, priority }: SmsMessage): Promise<SmsResponse> {
        return this.sendSms([{ number, message, senderid: senderid || this.senderId, priority }]);
    }

    /**
     * Sends bulk SMS messages.
     * @param messages Array of SmsMessage
     */
    async sendBulkSms(messages: SmsMessage[]): Promise<SmsResponse> {
        return this.sendSms(messages);
    }
}
