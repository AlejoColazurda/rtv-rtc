import { Component, input, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Invitation } from '../../models/invitation';

@Component({
  selector: 'app-document',
  imports: [CommonModule],
  templateUrl: './document.html'
})
export class DocumentComponent {
  readonly invitation = input.required<Invitation>();

  // Computed style object for setting custom variables in real-time
  readonly docStyles = computed(() => {
    const inv = this.invitation();
    const l = inv.layout;
    
    let width = '210mm';
    let height = '297mm';
    
    if (l.pageSize === 'custom') {
      width = `${l.widthMm}mm`;
      height = `${l.heightMm}mm`;
    } else if (l.pageSize === 'A4') {
      width = '210mm';
      height = '297mm';
    } else if (l.pageSize === 'A5') {
      width = '148mm';
      height = '210mm';
    } else if (l.pageSize === 'Letter') {
      width = '216mm';
      height = '279mm';
    }

    return {
      '--doc-width': width,
      '--doc-height': height,
      '--doc-margin-top': `${l.marginTop || 10}mm`,
      '--doc-margin-right': `${l.marginRight || 10}mm`,
      '--doc-margin-bottom': `${l.marginBottom || 10}mm`,
      '--doc-margin-left': `${l.marginLeft || 10}mm`,
      '--doc-font-scale': `${l.fontSizeScale || 1.0}`,
      '--col-qty-w': `${l.colQtyWidth || 15}%`,
      '--col-desc-w': `${l.colDescWidth || 50}%`,
      '--col-price-w': `${l.colPriceWidth || 18}%`,
      '--col-total-w': `${l.colTotalWidth || 17}%`
    };
  });
  
  // Calculate total amount
  readonly totalAmount = computed(() => {
    const inv = this.invitation();
    if (!inv || !inv.items) return 0;
    return inv.items.reduce((sum, item) => sum + (item.qty * item.unitPrice), 0);
  });

  formatDate(dateStr: string): string {
    if (!dateStr) return '';
    try {
      const d = new Date(dateStr);
      return d.toLocaleString('es-AR', {
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

  formatSimpleDate(dateStr: string): string {
    if (!dateStr) return '';
    try {
      const d = new Date(dateStr);
      return d.toLocaleDateString('es-AR', {
        day: '2-digit',
        month: '2-digit',
        year: 'numeric'
      });
    } catch {
      return dateStr;
    }
  }
}
