import { Component, inject, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Router } from '@angular/router';
import { Invitation } from '../../models/invitation';
import { InvitationService } from '../../services/invitation';

@Component({
  selector: 'app-dashboard',
  imports: [CommonModule],
  templateUrl: './dashboard.html'
})
export class DashboardComponent {
  private readonly invitationService = inject(InvitationService);
  private readonly router = inject(Router);

  // States
  readonly invitations = signal<Invitation[]>([]);
  readonly loading = signal(true);
  readonly error = signal<string | null>(null);

  // Computed metrics
  readonly stats = computed(() => {
    const list = this.invitations();
    return {
      total: list.length,
      accepted: list.filter(i => i.status === 'accepted').length,
      declined: list.filter(i => i.status === 'declined').length,
      pending: list.filter(i => i.status === 'pending' || !i.status).length
    };
  });

  constructor() {
    this.loadInvitations();
  }

  loadInvitations() {
    this.loading.set(true);
    this.invitationService.getInvitations().subscribe({
      next: (data) => {
        this.invitations.set(data || []);
        this.loading.set(false);
      },
      error: (err) => {
        console.error(err);
        this.error.set('No se pudo conectar con el servidor backend en Go. Asegúrate de que esté corriendo en el puerto 8080.');
        this.loading.set(false);
      }
    });
  }

  getSharedLink(id: string): string {
    const base = window.location.origin;
    return `${base}/invitation/${id}`;
  }

  copyLink(id: string) {
    const link = this.getSharedLink(id);
    navigator.clipboard.writeText(link).then(() => {
      alert('¡Enlace de invitación copiado!');
    }).catch(err => {
      console.error('Error copying text: ', err);
    });
  }

  downloadPDF(id: string) {
    window.open(this.invitationService.getPDFUrl(id), '_blank');
  }

  deleteInvitation(id: string) {
    if (confirm('¿Estás seguro de que deseas eliminar esta invitación? Esta acción no se puede deshacer.')) {
      this.invitationService.deleteInvitation(id).subscribe({
        next: () => {
          this.loadInvitations();
        },
        error: (err) => {
          console.error('Error deleting invitation:', err);
          alert('No se pudo eliminar la invitación. Revisa la conexión con el servidor.');
        }
      });
    }
  }

  navigateToCreator() {
    this.router.navigate(['/create']);
  }

  navigateToEditor(id: string) {
    this.router.navigate(['/edit', id]);
  }

  navigateToGuestView(id: string) {
    this.router.navigate(['/invitation', id]);
  }

  formatDate(dateStr: string): string {
    if (!dateStr) return '';
    try {
      const d = new Date(dateStr);
      return d.toLocaleDateString('es-AR', {
        day: '2-digit',
        month: '2-digit',
        year: 'numeric',
        hour: '2-digit',
        minute: '2-digit'
      });
    } catch {
      return dateStr;
    }
  }

  getMainConcept(inv: Invitation): string {
    if (inv.items && inv.items.length > 0) {
      return inv.items[0].description;
    }
    return 'Invitación de cortesía';
  }
}
