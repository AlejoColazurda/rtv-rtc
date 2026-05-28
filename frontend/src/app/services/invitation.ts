import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { Invitation } from '../models/invitation';

@Injectable({
  providedIn: 'root'
})
export class InvitationService {
  private readonly http = inject(HttpClient);
  private readonly apiUrl = 'http://localhost:8080/api/invitations';

  getInvitations(): Observable<Invitation[]> {
    return this.http.get<Invitation[]>(this.apiUrl);
  }

  getInvitation(id: string): Observable<Invitation> {
    return this.http.get<Invitation>(`${this.apiUrl}/${id}`);
  }

  createInvitation(invitation: Invitation): Observable<Invitation> {
    return this.http.post<Invitation>(this.apiUrl, invitation);
  }

  submitRSVP(id: string, status: 'accepted' | 'declined', signature: string, rejectionReason: string): Observable<Invitation> {
    return this.http.put<Invitation>(`${this.apiUrl}/${id}/rsvp`, {
      status,
      signature,
      rejectionReason
    });
  }

  updateInvitation(id: string, invitation: Invitation): Observable<Invitation> {
    return this.http.put<Invitation>(`${this.apiUrl}/${id}`, invitation);
  }

  deleteInvitation(id: string): Observable<any> {
    return this.http.delete<any>(`${this.apiUrl}/${id}`);
  }

  getPDFUrl(id: string): string {
    return `${this.apiUrl}/${id}/pdf`;
  }
}
