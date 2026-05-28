import { Component, ElementRef, ViewChild, inject, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, Router } from '@angular/router';
import { Invitation } from '../../models/invitation';
import { InvitationService } from '../../services/invitation';
import { DocumentComponent } from '../document/document';

@Component({
  selector: 'app-guest-view',
  imports: [CommonModule, FormsModule, DocumentComponent],
  templateUrl: './guest-view.html'
})
export class GuestViewComponent {
  private readonly route = inject(ActivatedRoute);
  private readonly router = inject(Router);
  private readonly invitationService = inject(InvitationService);

  @ViewChild('sigCanvas') canvasRef!: ElementRef<HTMLCanvasElement>;

  // States
  readonly invitation = signal<Invitation | null>(null);
  readonly loading = signal(true);
  readonly error = signal<string | null>(null);

  // Invitation Envelope State
  readonly envelopeOpen = signal(false);

  // RSVP Form States
  readonly showSignatureModal = signal(false);
  readonly showDeclineModal = signal(false);
  readonly submitting = signal(false);

  // Rejection Options
  rejectionReasonOption = 'cafe';
  customRejectionReason = '';

  // Canvas drawing properties
  private isDrawing = false;
  private ctx: CanvasRenderingContext2D | null = null;

  constructor() {
    this.loadInvitation();
  }

  loadInvitation() {
    const id = this.route.snapshot.paramMap.get('id');
    if (!id) {
      this.error.set('Identificador de invitación no válido.');
      this.loading.set(false);
      return;
    }

    this.invitationService.getInvitation(id).subscribe({
      next: (data) => {
        this.invitation.set(data);
        this.loading.set(false);
        // If invitation is already signed or declined, open the envelope directly
        if (data.status === 'accepted' || data.status === 'declined') {
          this.envelopeOpen.set(true);
        }
      },
      error: (err) => {
        console.error(err);
        this.error.set('No se pudo encontrar el remito solicitado. Verifica el enlace o el servidor.');
        this.loading.set(false);
      }
    });
  }

  toggleEnvelope() {
    // Only toggle if pending. If already accepted/declined, keep it open.
    const inv = this.invitation();
    if (inv && (inv.status === 'accepted' || inv.status === 'declined')) {
      return;
    }
    this.envelopeOpen.update(v => !v);
  }

  // RSVP: Accept flow
  openSignaturePad() {
    this.showSignatureModal.set(true);
    // Use setTimeout to ensure the canvas is rendered before accessing it
    setTimeout(() => this.initCanvas(), 100);
  }

  closeSignaturePad() {
    this.showSignatureModal.set(false);
  }

  // Canvas initialization and events
  initCanvas() {
    const canvas = this.canvasRef.nativeElement;
    // Set display resolution to match parent container size
    canvas.width = canvas.parentElement?.clientWidth || 400;
    canvas.height = 150;

    const context = canvas.getContext('2d');
    if (context) {
      context.strokeStyle = '#000080'; // Dark blue "conforme" ink
      context.lineWidth = 2.5;
      context.lineCap = 'round';
      context.lineJoin = 'round';
      this.ctx = context;
    }
  }

  clearCanvas() {
    if (this.ctx && this.canvasRef) {
      const canvas = this.canvasRef.nativeElement;
      this.ctx.clearRect(0, 0, canvas.width, canvas.height);
    }
  }

  // Mouse & Touch events mapping
  getCoordinates(event: MouseEvent | TouchEvent): { x: number; y: number } {
    const canvas = this.canvasRef.nativeElement;
    const rect = canvas.getBoundingClientRect();
    
    if (window.TouchEvent && event instanceof TouchEvent) {
      if (event.touches.length > 0) {
        return {
          x: event.touches[0].clientX - rect.left,
          y: event.touches[0].clientY - rect.top
        };
      }
    } else if (event instanceof MouseEvent) {
      return {
        x: event.clientX - rect.left,
        y: event.clientY - rect.top
      };
    }
    return { x: 0, y: 0 };
  }

  startDrawing(event: MouseEvent | TouchEvent) {
    event.preventDefault();
    this.isDrawing = true;
    const coords = this.getCoordinates(event);
    if (this.ctx) {
      this.ctx.beginPath();
      this.ctx.moveTo(coords.x, coords.y);
    }
  }

  draw(event: MouseEvent | TouchEvent) {
    if (!this.isDrawing || !this.ctx) return;
    event.preventDefault();
    const coords = this.getCoordinates(event);
    this.ctx.lineTo(coords.x, coords.y);
    this.ctx.stroke();
  }

  stopDrawing() {
    this.isDrawing = false;
  }

  confirmRSVPAccepted() {
    const inv = this.invitation();
    if (!inv || !inv.id) return;

    this.submitting.set(true);
    let signatureBase64 = '';

    if (this.canvasRef) {
      const canvas = this.canvasRef.nativeElement;
      // Convert to image URL
      signatureBase64 = canvas.toDataURL('image/png');
    }

    this.invitationService.submitRSVP(inv.id, 'accepted', signatureBase64, '').subscribe({
      next: (res) => {
        this.invitation.set(res);
        this.submitting.set(false);
        this.showSignatureModal.set(false);
        this.envelopeOpen.set(true);
        alert('¡Remito Recepcionado con éxito! Confirmaste tu asistencia.');
      },
      error: (err) => {
        console.error(err);
        this.submitting.set(false);
        alert('Ocurrió un error al firmar el remito.');
      }
    });
  }

  // RSVP: Decline flow
  openDeclineForm() {
    this.showDeclineModal.set(true);
  }

  closeDeclineForm() {
    this.showDeclineModal.set(false);
  }

  confirmRSVPDeclined() {
    const inv = this.invitation();
    if (!inv || !inv.id) return;

    this.submitting.set(true);
    let reason = '';
    
    if (this.rejectionReasonOption === 'cafe') {
      reason = 'Insuficiencia de cafeína crónica';
    } else if (this.rejectionReasonOption === 'dieta') {
      reason = 'Dieta estricta cero harinas/medialunas';
    } else if (this.rejectionReasonOption === 'reunion') {
      reason = 'Reunión improductiva de urgencia';
    } else {
      reason = this.customRejectionReason || 'No disponible';
    }

    this.invitationService.submitRSVP(inv.id, 'declined', '', reason).subscribe({
      next: (res) => {
        this.invitation.set(res);
        this.submitting.set(false);
        this.showDeclineModal.set(false);
        this.envelopeOpen.set(true);
        alert('Rechazaste la mercadería (Invitación rechazada).');
      },
      error: (err) => {
        console.error(err);
        this.submitting.set(false);
        alert('Ocurrió un error al enviar el rechazo.');
      }
    });
  }

  downloadPDF() {
    const inv = this.invitation();
    if (inv && inv.id) {
      window.open(this.invitationService.getPDFUrl(inv.id), '_blank');
    }
  }

  goToDashboard() {
    this.router.navigate(['/']);
  }
}
